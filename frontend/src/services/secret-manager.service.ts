import { z } from 'zod';

export const serviceAccountKeySchema = z.object({
    type: z.literal('service_account'),
    project_id: z.string(),
    private_key_id: z.string(),
    private_key: z.string(),
    client_email: z.email(),
    client_id: z.string(),
    auth_uri: z.url(),
    token_uri: z.url(),
    auth_provider_x509_cert_url: z.url().optional(),
    client_x509_cert_url: z.url().optional(),
    universe_domain: z.string().optional()
});

export type ServiceAccountKey = z.infer<typeof serviceAccountKeySchema>;

export class SecretManagerService {

    private static async importPrivateKey(pem: string): Promise<CryptoKey> {
        const pemHeader = "-----BEGIN PRIVATE KEY-----";
        const pemFooter = "-----END PRIVATE KEY-----";
        const pemContents = pem.substring(
            pem.indexOf(pemHeader) + pemHeader.length,
            pem.indexOf(pemFooter)
        ).replace(/\s/g, '');

        const binaryString = window.atob(pemContents);
        const binaryLen = binaryString.length;
        const bytes = new Uint8Array(binaryLen);

        for (let i = 0; i < binaryLen; i++) {
            bytes[i] = binaryString.charCodeAt(i);
        }

        return await window.crypto.subtle.importKey(
            "pkcs8",
            bytes.buffer,
            {
                name: "RSASSA-PKCS1-v1_5",
                hash: { name: "SHA-256" },
            },
            false,
            ["sign"]
        );
    }

    private static async createJWT(serviceAccount: ServiceAccountKey): Promise<string> {
        const now = Math.floor(Date.now() / 1000);
        const iat = now;
        const exp = now + 3600;

        const header = {
            alg: "RS256",
            typ: "JWT",
            kid: serviceAccount.private_key_id
        };

        const payload = {
            iss: serviceAccount.client_email,
            scope: "https://www.googleapis.com/auth/cloud-platform",
            aud: serviceAccount.token_uri,
            exp: exp,
            iat: iat
        };

        const base64UrlEncode = (input: string | Uint8Array) => {
            let binary = '';
            const bytes: Uint8Array = typeof input === 'string' ? new TextEncoder().encode(input) : input


            // Convert bytes to binary string
            const len = bytes.byteLength;
            for (let i = 0; i < len; i++) {
                binary += String.fromCharCode(bytes[i]);
            }

            // Base64 encode and make URL safe
            return window.btoa(binary)
                .replace(/\+/g, '-')
                .replace(/\//g, '_')
                .replace(/=+$/, '');
        };

        const safeHeader = base64UrlEncode(JSON.stringify(header));
        const safePayload = base64UrlEncode(JSON.stringify(payload));

        const dataToSign = `${safeHeader}.${safePayload}`;

        const key = await this.importPrivateKey(serviceAccount.private_key);
        const signature = await window.crypto.subtle.sign(
            "RSASSA-PKCS1-v1_5",
            key,
            new TextEncoder().encode(dataToSign)
        );

        const safeSignature = base64UrlEncode(new Uint8Array(signature));

        return `${safeHeader}.${safePayload}.${safeSignature}`;
    }

    public static async getAccessToken(serviceAccount: ServiceAccountKey): Promise<string> {
        const jwt = await this.createJWT(serviceAccount);

        const params = new URLSearchParams();
        params.append('grant_type', 'urn:ietf:params:oauth:grant-type:jwt-bearer');
        params.append('assertion', jwt);

        const response = await fetch(serviceAccount.token_uri, {
            method: 'POST',
            body: params
        });

        if (!response.ok) {
            const err = await response.text();
            throw new Error(`Failed to get access token: ${err}`);
        }

        const data = await response.json();
        return data.access_token;
    }

    public static async listSecretsWithToken(accessToken: string, projectId: string): Promise<string[]> {
        let secrets: string[] = [];
        let nextPageToken: string | undefined = undefined;

        do {
            const url = new URL(`https://secretmanager.googleapis.com/v1/projects/${projectId}/secrets`);
            url.searchParams.append('pageSize', '100');
            if (nextPageToken) {
                url.searchParams.append('pageToken', nextPageToken);
            }

            const response = await fetch(url.toString(), {
                headers: {
                    'Authorization': `Bearer ${accessToken}`
                }
            });

            if (!response.ok) {
                const err = await response.text();
                throw new Error(`Failed to list secrets: ${err}`);
            }

            const data = await response.json();
            if (data.secrets) {
                const names = data.secrets.map((s: any) => {
                    const parts = s.name.split('/');
                    return parts[parts.length - 1];
                });
                secrets = secrets.concat(names);
            }

            nextPageToken = data.nextPageToken;

        } while (nextPageToken);

        return secrets;
    }

    public static async listSecrets(serviceAccountKeyJson: string): Promise<string[]> {
        const serviceAccount: ServiceAccountKey = JSON.parse(serviceAccountKeyJson);
        const accessToken = await this.getAccessToken(serviceAccount);
        return this.listSecretsWithToken(accessToken, serviceAccount.project_id);
    }

    public static async testConnectionWithToken(accessToken: string, projectId: string): Promise<boolean> {
        try {
            // Fetch with pageSize=1 for a quick check
            const response = await fetch(`https://secretmanager.googleapis.com/v1/projects/${projectId}/secrets?pageSize=1`, {
                headers: {
                    'Authorization': `Bearer ${accessToken}`
                }
            });

            return response.ok;
        } catch (e) {
            console.error(e);
            return false;
        }
    }

    public static async testConnection(serviceAccountKeyJson: string): Promise<boolean> {
        try {
            const serviceAccount: ServiceAccountKey = JSON.parse(serviceAccountKeyJson);
            const accessToken = await this.getAccessToken(serviceAccount);
            return this.testConnectionWithToken(accessToken, serviceAccount.project_id);
        } catch (e) {
            console.error(e);
            return false;
        }
    }

    public static async getSecretValueWithToken(accessToken: string, projectId: string, secretName: string, version: string = 'latest'): Promise<{ value: string; version: string }> {
        const response = await fetch(`https://secretmanager.googleapis.com/v1/projects/${projectId}/secrets/${secretName}/versions/${version}:access`, {
            headers: {
                'Authorization': `Bearer ${accessToken}`
            }
        });

        if (!response.ok) {
            const err = await response.text();
            throw new Error(`Failed to fetch secret value: ${err}`);
        }

        const data = await response.json();
        // data.payload.data is base64 encoded
        const decodedValue = window.atob(data.payload.data);

        // Extract version from resource name "projects/*/secrets/*/versions/*"
        const nameParts = data.name.split('/');
        const versionId = nameParts[nameParts.length - 1];

        return {
            value: decodedValue,
            version: versionId
        };
    }

    public static async getSecretValue(serviceAccountKeyJson: string, secretName: string, version: string = 'latest'): Promise<{ value: string; version: string }> {
        const serviceAccount: ServiceAccountKey = JSON.parse(serviceAccountKeyJson);
        const accessToken = await this.getAccessToken(serviceAccount);
        return this.getSecretValueWithToken(accessToken, serviceAccount.project_id, secretName, version);
    }
}
