# Privy

A minimalist, peer-to-peer (P2P) encrypted chat application written in Go. It establishes a secure channel between two nodes using Elliptic-Curve Diffie-Hellman (ECDH) for key exchange and AES-GCM for symmetric message encryption, allowing custom IP/port configuration and dynamic Short Authentication String (SAS) generation.

## Features

- **P2P Architecture**: Direct communication between Host and Client over TCP across local networks or localhost.
- **Dynamic Connection Inputs**: Flexible configuration prompting the client for target Host IP addresses and custom ports.
- **Input Validation**: Automatic bounds checking ensuring port values fall strictly within the safe unprivileged range (1024-65535).
- **Dynamic Key Exchange**: X25519 Elliptic-Curve Diffie-Hellman (ECDH) handshake creates a unique session key per connection.
- **Out-of-Band Verification (SAS)**: MitM mitigation via a formatted Short Authentication String derived from a SHA-512 hash of the shared secret.
- **End-to-End Encryption**: Authenticated encryption using AES-GCM with unique nonces for every transmitted message.
- **Text-Safe Streaming**: Network payloads are encoded in Base64 to ensure transport stability.
- **Concurrent I/O**: Asynchronous message reading and writing managed through Go goroutines.

## Security Model: TOFU & SAS

The application implements a Trust on First Use (TOFU) model for seamless connectivity. To protect users against Man-in-the-Middle (MitM) attacks, the application prints a distinct, chunked Short Authentication String (SAS) derived from the derived shared key. Users can manually compare this code via an independent out-of-band secure channel (e.g., voice call) to verify identity authenticity.

## Network Port Validation Rule

The application implements a validation loop rejecting any ports out of the safe spectrum:
- **Valid Range**: 1024 to 65535 (Unprivileged/Private Ports).
- **Restricted Range**: 1 to 1023 (System/Well-Known Ports requiring root privileges are blocked to minimize runtime socket access errors).

## Architecture Workflow

1. **Connection Setup**: One node acts as Host (binding to a user-defined port) and the other connects as Client by inputting the specific target IP and matching port.
2. **Handshake Phase**: Both nodes generate an ephemeral X25519 key pair, exchange public keys in Base64 via the unified connection scanner, and locally derive the identical 32-byte shared secret.
3. **Authentication & Session**: The shared secret is hashed using SHA-512 to display the structured SAS code on both terminals. The derived secret is then loaded into an AES-GCM cipher block to encrypt and decrypt all subsequent text messages asynchronously.

## Prerequisites

- [Go 1.20 or higher](https://go.dev/doc/install)

## Getting Started

### Installation

Clone the repository and navigate into the project directory:

```bash
git clone https://github.com/bosioF/Privy.git
cd privy

```

### Usage

Run the application in two different terminal instances on your local machine or two machines connected to the same network.

#### Node A (Host)

Run the program and select the host option, then enter a valid port:

```bash
go run main.go

```

Prompt flow:

```text
You want to host(h) or connect(c)?
h
On what port do you want to listen? (1024-65535) 5000
Listening

```

#### Node B (Client)

Run the program, select the connect option, enter the target port, and provide the Host IP address (or press Enter for localhost):

```bash
go run main.go

```

Prompt flow:

```text
You want to host(h) or connect(c)?
c
On what port do you want to connect? (1024-65535) 5000
What is the IP? (Press Enter for localhost): 
Connecting to 127.0.0.1:5000...

```

Once the handshake is computed, the verification prompt will display on both terminals:

```text
Connection successful! 127.0.0.1:5000
Handshake successful! Connection Secure!
If you suspect someone is attempting a MitM attack, verify that this code is the same as the other person, over another secure channel.
SAS Code:  a1b2-c3d4-e5f6-7g8h 

```

You can now type messages in the terminal and press Enter to stream encrypted data.

## Cryptographic Specifications

* **Key Exchange**: Curve25519 (crypto/ecdh)
* **Authentication Hashing**: SHA-512 (crypto/sha512)
* **Symmetric Cipher**: AES-256-GCM (crypto/aes, crypto/cipher)
* **Payload Encoding**: Standard Base64 (encoding/base64)
