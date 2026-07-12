# Privy Chat

A minimalist, peer-to-peer (P2P) encrypted chat application written in Go. It establishes a secure channel between two nodes using Elliptic-Curve Diffie-Hellman (ECDH) for key exchange and AES-GCM for symmetric message encryption, allowing custom port configuration with built-in validation checks.

## Features

- **P2P Architecture**: Direct communication between Host and Client over TCP.
- **Dynamic Port Selection**: Users can input custom ports to run or connect to the service.
- **Input Validation**: Automatic bounds checking ensuring port values fall strictly within the safe unprivileged range (1024-65535).
- **Dynamic Key Exchange**: X25519 Elliptic-Curve Diffie-Hellman (ECDH) handshake creates a unique session key per connection.
- **End-to-End Encryption**: Authenticated encryption using AES-GCM with unique nonces for every transmitted message.
- **Text-Safe Streaming**: Network payloads are encoded in Base64 to ensure transport stability.
- **Concurrent I/O**: Asynchronous message reading and writing managed through Go goroutines.

## Network Port Validation Rule

The application implements a validation loop rejecting any ports out of the safe spectrum:
- **Valid Range**: 1024 to 65535 (Unprivileged/Private Ports).
- **Restricted Range**: 1 to 1023 (System/Well-Known Ports requiring root privileges are blocked to minimize runtime socket access errors).

## Architecture Workflow

1. **Connection Setup**: One node acts as Host (binding to a user-defined port) and the other connects as Client to the matching local network socket.
2. **Handshake Phase**: Both nodes generate an ephemeral X25519 key pair, exchange their public keys in Base64 via the unified connection scanner, and locally derive the identical 32-byte shared secret.
3. **Secure Chat Session**: The ephemeral keys are discarded. The derived shared secret is loaded into an AES-GCM cipher block to encrypt and decrypt all subsequent text messages asynchronously.

## Prerequisites

- Go 1.20 or higher

## Getting Started

### Installation

Clone the repository and navigate into the project directory:

```bash
git clone [https://github.com/bosioF/privy.git](https://github.com/bosioF/privy.git)
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

Run the program and select the connect option, entering the same target port chosen by the host:

```bash
go run main.go

```

Prompt flow:

```text
You want to host(h) or connect(c)?
c
On what port do you want to connect? (1024-65535) 5000
Connection successful! 127.0.0.1:5000

```

Once the handshake confirmation `Handshake successful! Connection Secure!` appears on both screens, you can type messages in the terminal and press Enter to stream encrypted data.

## Cryptographic Specifications

* **Key Exchange**: Curve25519 (crypto/ecdh)
* **Symmetric Cipher**: AES-256-GCM (crypto/aes, crypto/cipher)
* **Payload Encoding**: Standard Base64 (encoding/base64)
