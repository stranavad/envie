import {x25519} from '@noble/curves/ed25519.js';
import {sha256} from "@noble/hashes/sha2.js";


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

    /**
     * Generate a random AES key for project symmetric encryption
     */
    static generateAesKey(): string {
        const key = crypto.getRandomValues(new Uint8Array(32));
        return bytesToBase64(key);
    }

    /**
     * AES-GCM Symmetric Encryption
     * Matches Rust: [Nonce (12)][Ciphertext]
     */
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

    /**
     * AES-GCM Symmetric Decryption
     * Expects: [Nonce (12)][Ciphertext]
     */
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

    /**
     * Asymmetric Encryption (ECDH + AES-GCM)
     * Matches Rust logic:
     * 1. Ephemeral Key Pair
     * 2. ECDH Shared Secret (Ephemeral Priv + Recipient Pub)
     * 3. HKDF/Hash (SHA256) -> Single AES Key
     * 4. Encrypt Payload
     * 5. Pack: [Ephemeral Public (32)][Nonce (12)][Ciphertext]
     */
    static async encryptKey(recipientPublicKeyB64: string, inputDataB64: string): Promise<string> {
        const recipientPub = base64ToBytes(recipientPublicKeyB64);

        // Input payload is expected to be Base64 string in the Rust version? 
        // "input_data" in Rust is decoded from Base64 before encryption.
        // So we need to decode it first to get the raw bytes to encrypt.
        const payload = base64ToBytes(inputDataB64);

        // 1. Ephemeral Key
        const ephemeralPriv = x25519.utils.randomSecretKey();
        const ephemeralPub = x25519.getPublicKey(ephemeralPriv);

        // 2. Shared Secret
        const sharedSecret = x25519.getSharedSecret(ephemeralPriv, recipientPub);

        // 3. Derive Key (SHA256 hash of shared secret, as per Rust impl)
        // Rust: hasher.update(shared_secret); let derived = hasher.finalize();
        const derivedKeyBytes = sha256(sharedSecret);

        // 4. Encrypt with AES-GCM
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

        // 5. Pack
        const combined = concatBytes(ephemeralPub, iv, ciphertext);
        return bytesToBase64(combined);
    }

    /**
     * Asymmetric Decryption
     * Expects: [Ephemeral Public (32)][Nonce (12)][Ciphertext]
     */
    static async decryptKey(privateKeyB64: string, encryptedDataB64: string): Promise<string> {
        const privKey = base64ToBytes(privateKeyB64);
        const combined = base64ToBytes(encryptedDataB64);

        if (combined.length < 32 + 12) throw new Error("Data too short");

        const ephemeralPub = combined.slice(0, 32);
        const iv = combined.slice(32, 32 + 12);
        const ciphertext = combined.slice(32 + 12);

        // 1. Shared Secret
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
}
