import * as bip39 from '@scure/bip39';
import { sha256 } from "@noble/hashes/sha2.js";
import { wordlist } from "@scure/bip39/wordlists/english.js";
import { x25519 } from '@noble/curves/ed25519.js';

export class IdentityService {
    private static MASTER_KEY_MEMORY: string | null = null; // Stored in memory only while app is running

    /**
     * Generates a new random 12-word recovery phrase (128-bit entropy)
     */
    static generateRecoveryPhrase(): string {
        return bip39.generateMnemonic(wordlist, 128);
    }

    /**
     * Validates a recovery phrase
     */
    static validateRecoveryPhrase(phrase: string): boolean {
        return bip39.validateMnemonic(phrase, wordlist);
    }

    /**
     * Derives the Master User Key (32 bytes / Base64) from the recovery phrase.
     * Uses PBKDF2 (via bip39.mnemonicToSeed) -> SHA256 to get fixed 32 bytes.
     */
    static async deriveMasterKey(phrase: string): Promise<string> {
        // 1. Get Seed (64 bytes usually)
        const seed = await bip39.mnemonicToSeed(phrase, "");

        // 2. Hash seed to get 32-byte key for AES/X25519 consumption
        const keyBytes = sha256(seed);


        // Using our helper from EncryptionService would be better, but let's inline simple implementation or make helper public
        return this.bytesToBase64(keyBytes);
    }

    /**
     * Sets the Master Key in memory (after login/recovery)
     */
    static setMasterKey(keyB64: string) {
        this.MASTER_KEY_MEMORY = keyB64;
    }

    static getMasterKey(): string | null {
        return this.MASTER_KEY_MEMORY;
    }

    /**
     * Clears sensitive data from memory
     */
    static clear() {
        this.MASTER_KEY_MEMORY = null;
    }

    private static bytesToBase64(bytes: Uint8Array): string {
        const binString = Array.from(bytes, (byte) => String.fromCodePoint(byte)).join("");
        return btoa(binString);
    }

    private static base64ToBytes(base64: string): Uint8Array {
        const binString = atob(base64);
        return Uint8Array.from(binString, (m) => m.codePointAt(0)!);
    }

    /**
     * treat the symmetric 32-byte Master Key as a Private Key seed
     * and return the KeyPair (Public + Private) for asymmetric operations.
     */
    static getMasterKeyPair(): { publicKey: string, privateKey: string } | null {
        if (!this.MASTER_KEY_MEMORY) return null;

        try {
            // 1. Decode Base64 Master Key
            const privateBytes = this.base64ToBytes(this.MASTER_KEY_MEMORY);

            // 2. Derive Public Key (X25519)
            const publicBytes = x25519.getPublicKey(privateBytes);

            return {
                privateKey: this.MASTER_KEY_MEMORY,
                publicKey: this.bytesToBase64(publicBytes)
            };
        } catch (e) {
            console.error("Failed to derive master key pair", e);
            return null;
        }
    }
}
