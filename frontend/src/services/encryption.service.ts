import {x25519} from '@noble/curves/ed25519.js';
import {sha256} from "@noble/hashes/sha2.js";
import {hkdf} from "@noble/hashes/hkdf.js";


export interface KeyPair {
    privateKey: string;
    publicKey: string;
}

function base64ToBytes(base64: string): Uint8Array {
    const binString = atob(base64);
    return Uint8Array.from(binString, (m) => m.codePointAt(0)!);
}

function bytesToBase64(bytes: Uint8Array): string {
    const binString = Array.from(bytes, (byte) => String.fromCodePoint(byte)).join("");
    return btoa(binString);
}

function concatBytes(...arrays: Uint8Array[]): Uint8Array {
    let totalLength = 0;
    for (const arr of arrays) {
        totalLength += arr.length;
    }
    const result = new Uint8Array(totalLength);
    let offset = 0;
    for (const arr of arrays) {
        result.set(arr, offset);
        offset += arr.length;
    }
    return result;
}

export class EncryptionService {
    /**
     * Generate key pair (public + private) for user encryption (asymmetric)
     */
    static generateKeyPair(): KeyPair {
        const privateKey = x25519.utils.randomSecretKey();
        const publicKey = x25519.getPublicKey(privateKey);

        return {
            privateKey: bytesToBase64(privateKey),
            publicKey: bytesToBase64(publicKey)
        };
    }


    static generateAesKey(): string {
        const key = crypto.getRandomValues(new Uint8Array(32));
        return bytesToBase64(key);
    }


    static async encryptValue(keyB64: string, plaintext: string): Promise<string> {
        const keyBytes = base64ToBytes(keyB64);
        if (keyBytes.length !== 32) throw new Error("Invalid key length");

        // Import Raw Key for AES-GCM
        const key = await crypto.subtle.importKey(
            "raw",
            keyBytes,
            "AES-GCM",
            false,
            ["encrypt"]
        );

        const iv = crypto.getRandomValues(new Uint8Array(12)); // Nonce (12 bytes)
        const encBuilder = new TextEncoder();
        const encodedPlaintext = encBuilder.encode(plaintext);

        const ciphertextBuffer = await crypto.subtle.encrypt(
            {
                name: "AES-GCM",
                iv: iv
            },
            key,
            encodedPlaintext
        );

        const ciphertext = new Uint8Array(ciphertextBuffer);
        const combined = concatBytes(iv, ciphertext);

        return bytesToBase64(combined);
    }

    static async decryptValue(keyB64: string, combinedB64: string): Promise<string> {
        const keyBytes = base64ToBytes(keyB64);
        if (keyBytes.length !== 32) throw new Error("Invalid key length");

        const combined = base64ToBytes(combinedB64);
        if (combined.length < 12) throw new Error("Ciphertext too short");

        const iv = combined.slice(0, 12);
        const ciphertext = combined.slice(12);

        const key = await crypto.subtle.importKey(
            "raw",
            keyBytes,
            "AES-GCM",
            false,
            ["decrypt"]
        );

        const plaintextBuffer = await crypto.subtle.decrypt(
            {
                name: "AES-GCM",
                iv: iv
            },
            key,
            ciphertext
        );

        const decBuilder = new TextDecoder();
        return decBuilder.decode(plaintextBuffer);
    }


    static async encryptKey(recipientPublicKeyB64: string, inputDataB64: string): Promise<string> {
        const recipientPub = base64ToBytes(recipientPublicKeyB64);
        const payload = base64ToBytes(inputDataB64);

        // 1. Ephemeral Key
        const ephemeralPriv = x25519.utils.randomSecretKey();
        const ephemeralPub = x25519.getPublicKey(ephemeralPriv);

        // 2. Shared Secret
        const sharedSecret = x25519.getSharedSecret(ephemeralPriv, recipientPub);

        const derivedKeyBytes = sha256(sharedSecret);

        const key = await crypto.subtle.importKey(
            "raw",
            derivedKeyBytes,
            "AES-GCM",
            false,
            ["encrypt"]
        );

        const iv = crypto.getRandomValues(new Uint8Array(12));

        const ciphertextBuffer = await crypto.subtle.encrypt(
            {
                name: "AES-GCM",
                iv: iv
            },
            key,
            payload
        );

        const ciphertext = new Uint8Array(ciphertextBuffer);

        const combined = concatBytes(ephemeralPub, iv, ciphertext);
        return bytesToBase64(combined);
    }

    static async decryptKey(privateKeyB64: string, encryptedDataB64: string): Promise<string> {
        const privKey = base64ToBytes(privateKeyB64);
        const combined = base64ToBytes(encryptedDataB64);

        if (combined.length < 32 + 12) throw new Error("Data too short");

        const ephemeralPub = combined.slice(0, 32);
        const iv = combined.slice(32, 32 + 12);
        const ciphertext = combined.slice(32 + 12);

        const sharedSecret = x25519.getSharedSecret(privKey, ephemeralPub);

        // 2. Derive Key
        const derivedKeyBytes = sha256(sharedSecret);

        // 3. Decrypt
        const key = await crypto.subtle.importKey(
            "raw",
            derivedKeyBytes,
            "AES-GCM",
            false,
            ["decrypt"]
        );

        const plaintextBuffer = await crypto.subtle.decrypt(
            {
                name: "AES-GCM",
                iv: iv
            },
            key,
            ciphertext
        );

        // Rust returns this as Base64 encoded string
        return bytesToBase64(new Uint8Array(plaintextBuffer));
    }

    /**
     * Generate a new random project key
     * This is the key used to encrypt config items
     */
    static generateProjectKey(): string {
        return this.generateAesKey();
    }

    /**
     * Generate a CLI access token with derived cryptographic material
     */
    static async generateAccessToken(projectKeyB64: string): Promise<GeneratedAccessToken> {
        const tokenBytes = crypto.getRandomValues(new Uint8Array(32));
        const encoder = new TextEncoder();

        // base64 URL encoding without padding
        const encoded = bytesToBase64Url(tokenBytes);
        const token = 'envie_' + encoded;
        const tokenPrefix = encoded.substring(0, 3);

        // Derive identity ID (16 bytes)
        const identityIdBytes = hkdf(sha256, tokenBytes, undefined, encoder.encode('envie-identity-id'), 16);

        // Hash identity ID for server storage
        const identityIdHash = bytesToHex(sha256(identityIdBytes));

        // Derive private key (32 bytes)
        const privateKey = hkdf(sha256, tokenBytes, undefined, encoder.encode('envie-private-key'), 32);

        // Derive public key
        const publicKey = x25519.getPublicKey(privateKey);

        // Encrypt project key to the token's public key
        const encryptedProjectKey = await this.encryptKeyToPublicKey(publicKey, projectKeyB64);

        return {
            token,
            tokenPrefix,
            identityIdHash,
            encryptedProjectKey,
        };
    }

    private static async encryptKeyToPublicKey(publicKey: Uint8Array, keyB64: string): Promise<string> {
        const payload = base64ToBytes(keyB64);
        const encoder = new TextEncoder();

        const ephemeralPriv = x25519.utils.randomSecretKey();
        const ephemeralPub = x25519.getPublicKey(ephemeralPriv);

        const sharedSecret = x25519.getSharedSecret(ephemeralPriv, publicKey);

        // Derive AES key using HKDF (matching CLI's "envie-encrypt" info)
        const derivedKey = hkdf(sha256, sharedSecret, undefined, encoder.encode('envie-encrypt'), 32);

        const key = await crypto.subtle.importKey('raw', derivedKey, 'AES-GCM', false, ['encrypt']);
        const iv = crypto.getRandomValues(new Uint8Array(12));
        const ciphertextBuffer = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, payload);
        const ciphertext = new Uint8Array(ciphertextBuffer);

        const combined = concatBytes(ephemeralPub, iv, ciphertext);
        return bytesToBase64(combined);
    }
}

export interface GeneratedAccessToken {
    token: string;
    tokenPrefix: string;
    identityIdHash: string;
    encryptedProjectKey: string;
}

function bytesToBase64Url(bytes: Uint8Array): string {
    const base64 = bytesToBase64(bytes);
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function bytesToHex(bytes: Uint8Array): string {
    return Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('');
}
