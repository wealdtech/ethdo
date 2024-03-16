#!/usr/bin/env python3
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.backends import default_backend
import hashlib
import base64
import json
import getpass

def create_crypto_key(password):
    # Hash the password using SHA-256 and take the first 32 characters of base64-encoded hash
    hashed_password = hashlib.sha256(password.encode()).digest()
    base64_encoded = base64.b64encode(hashed_password)
    return base64_encoded.decode()[:32]

def decrypt(crypto_key, value):
    # Ensure the crypto_key is 32 bytes for AES-256
    if len(crypto_key) != 32:
        raise ValueError("The crypto key must be 32 bytes long for AES-256.")

    # Decoding the value from hex to bytes
    encrypted_text = bytes.fromhex(value)

    # Create a Cipher object using the AES algorithm and ECB mode
    cipher = Cipher(algorithms.AES(crypto_key.encode()), modes.ECB(), backend=default_backend())

    # Decrypt the data
    decryptor = cipher.decryptor()
    decrypted = decryptor.update(encrypted_text) + decryptor.finalize()

    # Assuming the decrypted data is a base64 encoded string
    decoded_data = base64.b64decode(decrypted)

    # Convert to string and parse as JSON
    return json.loads(decoded_data.decode('ascii'))

encrypted_value = input("Enter the encrypted seed value: ")
password = getpass.getpass("Enter the password: ")
crypto_key = create_crypto_key(password)

try:
    decrypted_seed = decrypt(crypto_key, encrypted_value)
    print("Decrypted Seed:", decrypted_seed)
except Exception as e:
    print("Error during decryption:", str(e))
