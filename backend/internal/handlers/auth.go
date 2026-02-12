package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"envie-backend/internal/auth"
	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func AuthLogin(c *gin.Context) {
	authURL := auth.OAuthConfig.AuthCodeURL("", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func AuthLoginGoogle(c *gin.Context) {
	authURL := auth.GoogleOAuthConfig.AuthCodeURL("", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func AuthCallback(c *gin.Context) {
	code := c.Query("code")

	githubUser, err := auth.GetGithubUser(code)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "", renderErrorPage("Authentication failed: "+err.Error()))
		return
	}


	var user models.User
	if err := database.DB.Where("github_id = ?", githubUser.ID).First(&user).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			RespondInternalError(c, "Failed to check for user existence")
			return
		}

		user = models.User{
			Name:      githubUser.Name,
			Email:     githubUser.Email,
			AvatarURL: githubUser.AvatarURL,
			GithubID:  githubUser.ID,
			PublicKey: nil,
		}

		if user.Name == "" {
			user.Name = githubUser.Login
		}

		if err := database.DB.Create(&user).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "", renderErrorPage("Failed to create user: "+err.Error()))
			return
		}
	} else {
		user.Name = githubUser.Name
		user.Email = githubUser.Email
		user.AvatarURL = githubUser.AvatarURL

		if user.Name == "" {
			user.Name = githubUser.Login
		}

		database.DB.Save(&user)
	}

	// Clean old linking codes
	database.DB.Where("user_id = ? AND (used_at IS NOT NULL OR expires_at < ?)", user.ID, time.Now()).
		Delete(&models.LinkingCode{})

	linkingCode, err := auth.GenerateLinkingCode()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "", renderErrorPage("Failed to generate linking code"))
		return
	}

	linkingCodeRecord := models.LinkingCode{
		Code:      strings.ToUpper(linkingCode),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(auth.LinkingCodeDuration),
	}

	if err := database.DB.Create(&linkingCodeRecord).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "", renderErrorPage("Failed to save linking code: "+err.Error()))
		return
	}

	log.Printf("Created linking code for user %s: %s", user.ID, strings.ToUpper(linkingCode))

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, renderLinkingCodePage(strings.ToUpper(linkingCode), user.Name))
}

func AuthCallbackGoogle(c *gin.Context) {
	code := c.Query("code")

	googleUser, err := auth.GetGoogleUser(code)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "", renderErrorPage("Authentication failed: "+err.Error()))
		return
	}

	var user models.User
	// First try to find by Google ID
	err = database.DB.Where("google_id = ?", googleUser.ID).First(&user).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			RespondInternalError(c, "Failed to check for user existence")
			return
		}

		// Not found by Google ID — try by email (account linking)
		err = database.DB.Where("email = ?", googleUser.Email).First(&user).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				RespondInternalError(c, "Failed to check for user existence")
				return
			}

			// Brand new user
			user = models.User{
				Name:      googleUser.Name,
				Email:     googleUser.Email,
				AvatarURL: googleUser.AvatarURL,
				GoogleID:  googleUser.ID,
				PublicKey: nil,
			}

			if err := database.DB.Create(&user).Error; err != nil {
				c.HTML(http.StatusInternalServerError, "", renderErrorPage("Failed to create user: "+err.Error()))
				return
			}
		} else {
			// Existing user found by email — link Google ID
			user.GoogleID = googleUser.ID
			database.DB.Save(&user)
		}
	} else {
		// Found by Google ID — update profile
		user.Name = googleUser.Name
		user.Email = googleUser.Email
		user.AvatarURL = googleUser.AvatarURL
		database.DB.Save(&user)
	}

	// Clean old linking codes
	database.DB.Where("user_id = ? AND (used_at IS NOT NULL OR expires_at < ?)", user.ID, time.Now()).
		Delete(&models.LinkingCode{})

	linkingCode, err := auth.GenerateLinkingCode()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "", renderErrorPage("Failed to generate linking code"))
		return
	}

	linkingCodeRecord := models.LinkingCode{
		Code:      strings.ToUpper(linkingCode),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(auth.LinkingCodeDuration),
	}

	if err := database.DB.Create(&linkingCodeRecord).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "", renderErrorPage("Failed to save linking code: "+err.Error()))
		return
	}

	log.Printf("Created linking code for user %s: %s", user.ID, strings.ToUpper(linkingCode))

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, renderLinkingCodePage(strings.ToUpper(linkingCode), user.Name))
}

type ExchangeRequest struct {
	Code            string `json:"code" binding:"required"`
	DevicePublicKey string `json:"devicePublicKey"`
}

type ExchangeResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	User         struct {
		ID               uuid.UUID `json:"id"`
		Name             string    `json:"name"`
		Email            string    `json:"email"`
		AvatarURL        string    `json:"avatarUrl"`
		GithubID         int64     `json:"githubId"`
		GoogleID         string    `json:"googleId"`
		PublicKey        *string   `json:"publicKey"`
		MasterKeyVersion int       `json:"masterKeyVersion"`
	} `json:"user"`
}

func AuthExchange(c *gin.Context) {
	var req ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request: "+err.Error())
		return
	}

	code := strings.ToUpper(strings.TrimSpace(req.Code))

	var linkingCode models.LinkingCode
	if err := database.DB.Where("code = ?", code).First(&linkingCode).Error; err != nil {
		RespondUnauthorized(c, "Invalid or expired linking code")
		return
	}

	if !linkingCode.IsValid() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Linking code has expired or already been used"})
		return
	}

	now := time.Now()
	linkingCode.UsedAt = &now
	database.DB.Save(&linkingCode)

	var user models.User
	if err := database.DB.First(&user, "id = ?", linkingCode.UserID).Error; err != nil {
		RespondInternalError(c, "User not found")
		return
	}

	// Update device LastActive if device public key provided
	if req.DevicePublicKey != "" {
		database.DB.Model(&models.UserIdentity{}).
			Where("user_id = ? AND public_key = ?", user.ID, req.DevicePublicKey).
			Update("last_active", time.Now())
	}

	accessToken, err := auth.GenerateAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	response := ExchangeResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(auth.AccessTokenDuration.Seconds()),
	}
	response.User.ID = user.ID
	response.User.Name = user.Name
	response.User.Email = user.Email
	response.User.AvatarURL = user.AvatarURL
	response.User.GithubID = user.GithubID
	response.User.GoogleID = user.GoogleID
	response.User.PublicKey = user.PublicKey
	response.User.MasterKeyVersion = user.MasterKeyVersion

	c.JSON(http.StatusOK, response)
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func AuthRefresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate the refresh token JWT
	claims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Generate new tokens
	accessToken, err := auth.GenerateAccessToken(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	newRefreshToken, err := auth.GenerateRefreshToken(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": newRefreshToken,
		"expiresIn":    int(auth.AccessTokenDuration.Seconds()),
	})
}

func AuthLogout(c *gin.Context) {
	// With stateless JWT refresh tokens, we can't revoke server-side.
	// The client should discard the tokens locally.
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func renderLinkingCodePage(code string, userName string) string {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Envie - Authentication Successful</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: #09090b;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #fafafa;
        }
        .container {
            background: #09090b;
            border-radius: 8px;
            padding: 24px;
            text-align: center;
            width: 100%;
            max-width: 384px;
            border: 1px solid #27272a;
        }
        .logo {
            font-size: 30px;
            font-weight: 700;
            letter-spacing: -0.025em;
            margin-bottom: 4px;
            color: #fafafa;
        }
        .welcome {
            color: #a1a1aa;
            font-size: 14px;
            margin-bottom: 24px;
        }
        .success-icon {
            width: 48px;
            height: 48px;
            background: #18181b;
            border: 1px solid #27272a;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 16px;
        }
        .success-icon svg {
            width: 24px;
            height: 24px;
            stroke: #22c55e;
        }
        h1 {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 8px;
            color: #fafafa;
        }
        .instructions {
            color: #a1a1aa;
            margin-bottom: 24px;
            line-height: 1.5;
            font-size: 14px;
        }
        .code-container {
            background: #18181b;
            border: 1px solid #27272a;
            border-radius: 6px;
            padding: 16px;
            margin-bottom: 16px;
        }
        .code-label {
            font-size: 12px;
            font-weight: 500;
            color: #a1a1aa;
            margin-bottom: 8px;
        }
        .code {
            font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Monaco, 'Courier New', monospace;
            font-size: 24px;
            font-weight: 600;
            letter-spacing: 0.1em;
            color: #fafafa;
            user-select: all;
            cursor: pointer;
        }
        .expires {
            font-size: 12px;
            color: #71717a;
            margin-top: 8px;
        }
        .copy-btn {
            background: #fafafa;
            border: none;
            border-radius: 6px;
            padding: 10px 16px;
            color: #18181b;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            width: 100%;
            transition: opacity 0.2s;
        }
        .copy-btn:hover {
            opacity: 0.9;
        }
        .copy-btn:active {
            opacity: 0.8;
        }
        .copied {
            background: #22c55e !important;
            color: #fafafa !important;
        }
        .note {
            margin-top: 16px;
            padding-top: 16px;
            border-top: 1px solid #27272a;
            font-size: 12px;
            color: #71717a;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">Envie</div>
        <p class="welcome">Welcome back, {{.UserName}}</p>

        <div class="success-icon">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
            </svg>
        </div>

        <h1>Authentication Successful</h1>
        <p class="instructions">
            Copy the code below and paste it into the Envie app to complete sign-in.
        </p>

        <div class="code-container">
            <div class="code-label">Your linking code</div>
            <div class="code" id="code" onclick="copyCode()">{{.Code}}</div>
            <div class="expires">Expires in 5 minutes</div>
        </div>

        <button class="copy-btn" id="copyBtn" onclick="copyCode()">Copy Code</button>

        <p class="note">You can close this page after copying the code.</p>
    </div>

    <script>
        function copyCode() {
            const code = document.getElementById('code').textContent;
            navigator.clipboard.writeText(code).then(() => {
                const btn = document.getElementById('copyBtn');
                btn.textContent = 'Copied!';
                btn.classList.add('copied');
                setTimeout(() => {
                    btn.textContent = 'Copy Code';
                    btn.classList.remove('copied');
                }, 2000);
            });
        }
    </script>
</body>
</html>`

	t, _ := template.New("linkingCode").Parse(tmpl)
	var result strings.Builder
	t.Execute(&result, struct {
		Code     string
		UserName string
	}{Code: code, UserName: userName})
	return result.String()
}

func renderErrorPage(message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Envie - Error</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #09090b;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #fafafa;
        }
        .container {
            background: #09090b;
            border-radius: 8px;
            padding: 24px;
            text-align: center;
            width: 100%%;
            max-width: 384px;
            border: 1px solid #27272a;
        }
        .logo {
            font-size: 30px;
            font-weight: 700;
            letter-spacing: -0.025em;
            margin-bottom: 24px;
            color: #fafafa;
        }
        .error-icon {
            width: 48px;
            height: 48px;
            background: #18181b;
            border: 1px solid #27272a;
            border-radius: 50%%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 16px;
        }
        .error-icon svg {
            width: 24px;
            height: 24px;
            stroke: #ef4444;
        }
        h1 {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 8px;
            color: #fafafa;
        }
        .error {
            color: #a1a1aa;
            font-size: 14px;
            line-height: 1.5;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">Envie</div>
        <div class="error-icon">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
        </div>
        <h1>Authentication Failed</h1>
        <p class="error">%s</p>
    </div>
</body>
</html>`, message)
}
