# Go-GCSProxy Design Document

## 1. System Requirements

### 1.1 Hardware Requirements
- CPU: Minimum 2 cores recommended
- Memory: Minimum 4GB RAM
- Storage: 10GB minimum for proxy and certificate storage
- Network: 1Gbps minimum bandwidth

### 1.2 Software Requirements
- Operating System: Linux (recommended), macOS, or Windows
- Go version: 1.21 or later
- Docker: 20.10 or later (for containerized deployment)
- Google Cloud SDK: Latest version
- Required permissions:
  - Google Cloud KMS access
  - Google Cloud Storage access
  - IAM permissions for key management

### 1.3 Network Requirements
- Outbound access to Google Cloud APIs
- Inbound access on configured proxy port (default: 9080)
- SSL/TLS certificate management capabilities
- Support for HTTPS traffic interception

### 1.4 Performance Requirements
- Maximum file size: 10TB
- Concurrent connections: 1000+
- Encryption latency: < 100ms per operation
- Throughput: 1Gbps minimum
- Availability: 99.9% uptime

## 2. System Overview

Go-GCSProxy is an encrypting reverse proxy for Google Cloud Storage (GCS) that provides client-side encryption capabilities while maintaining compatibility with existing GCS tools and services.

### 2.1 Purpose
- Add an additional layer of security to GCS data storage
- Provide transparent encryption/decryption for GCS operations
- Support organizations with strict security and privacy requirements
- Maintain compatibility with existing GCS tools (gsutil, gcloud, axlearn, tensorflow)

### 2.2 Key Features
- Transparent encryption/decryption using Google Cloud KMS and Tink
- MITM proxy for HTTP/HTTPS traffic interception
- Per-bucket key management
- Compliance with approved cryptographic algorithms
- Scalable architecture with load balancer support
- Sidecar deployment capability
- Configurable logging with security best practices

## 3. Architecture

### 3.1 High-Level Architecture
```
[Client Applications] <---> [Go-GCSProxy] <---> [Google Cloud Storage]
        |                          |                    |
        |                          |                    |
        v                          v                    v
[gsutil/gcloud]            [Encryption Engine]    [GCS API]
        |                          |                    |
        |                          |                    |
        v                          v                    v
[Local Storage]            [KMS Integration]     [Cloud Storage]
```

### 3.2 Component Interaction
```
[Client Request] --> [Proxy Interception] --> [Encryption] --> [GCS Upload]
     ^                      |                      |              |
     |                      v                      v              v
     +---------------- [Decryption] <-------- [GCS Download] <----+
```

### 3.3 Deployment Architecture
```
[Load Balancer] --> [Go-GCSProxy Instances] --> [Google Cloud]
     |                      |                        |
     |                      |                        |
     v                      v                        v
[Client Traffic]    [Local Certificate Store]   [KMS/GCS APIs]
```

### 3.4 Core Components

#### 3.4.1 Proxy Server
- Man-in-the-Middle (MITM) proxy implementation
- HTTP/HTTPS traffic interception
- Request/response modification
- Certificate management
- OpenTelemetry integration for monitoring

#### 3.4.2 Encryption Engine
- Google Cloud KMS integration
- Tink library for cryptographic operations
- Per-bucket key mapping support
- Approved algorithms:
  - SHA-256
  - RSA-2048
  - ECDSA-256

#### 3.4.3 Configuration Management
- Environment variable based configuration
- Custom CA certificate support
- Debug level configuration
- KMS bucket key mapping

## 4. Data Flow

### 4.1 Upload Process
```
Client -> Proxy -> Encryption -> GCS
```
1. Client initiates upload request
2. Proxy intercepts request
3. Data is encrypted using appropriate KMS key
4. Encrypted data is forwarded to GCS
5. Metadata is stored for decryption

### 4.2 Download Process
```
GCS -> Proxy -> Decryption -> Client
```
1. Client initiates download request
2. Proxy intercepts request
3. Encrypted data is retrieved from GCS
4. Data is decrypted using appropriate KMS key
5. Decrypted data is returned to client

## 5. Implementation Details

### 5.1 Security Features
- Client-side encryption using Google Cloud KMS
- Per-bucket key mapping for granular control
- Secure certificate management
- Safe logging practices to prevent key/data leaks

### 5.2 Performance Considerations
- Support for large file uploads
- Streaming upload capability (planned)
- Efficient encryption/decryption operations
- Metrics collection for performance monitoring

### 5.3 Deployment Options
- Standalone service
- Docker container
- Sidecar deployment in GKE
- Load balancer support

## 6. Configuration

### 6.1 Environment Variables
- `GCP_KMS_RESOURCE_NAME`: KMS key resource path
- `PROXY_CERT_PATH`: Path to certificate files
- `SSL_INSECURE`: SSL verification settings
- `DEBUG_LEVEL`: Logging verbosity
- `GCP_KMS_BUCKET_KEY_MAPPING`: Bucket-to-key mappings

### 6.2 Client Configuration
- Proxy settings for gsutil/gcloud
- CA certificate trust configuration
- Environment variable setup

## 7. Error Handling and Recovery

### 7.1 Error Categories
1. **Configuration Errors**
   - Invalid KMS key configuration
   - Missing required environment variables
   - Invalid certificate configuration

2. **Runtime Errors**
   - KMS access failures
   - Encryption/decryption failures
   - GCS API errors
   - Network connectivity issues

3. **Resource Errors**
   - Memory exhaustion
   - File size limits exceeded
   - Concurrent connection limits

### 7.2 Error Handling Strategies
1. **Graceful Degradation**
   - Fallback to unencrypted mode when configured
   - Retry mechanisms for transient failures
   - Circuit breaker for repeated failures

2. **Error Reporting**
   - Structured logging with error codes
   - OpenTelemetry integration for monitoring
   - Alert thresholds for critical errors

3. **Recovery Procedures**
   - Automatic retry for transient failures
   - Manual intervention procedures
   - Backup and restore processes

### 7.3 Failure Scenarios
1. **KMS Unavailability**
   - Impact: Encryption/decryption operations fail
   - Mitigation: Circuit breaker, fallback options
   - Recovery: Automatic retry, manual key rotation

2. **Network Issues**
   - Impact: GCS operations fail
   - Mitigation: Connection pooling, timeouts
   - Recovery: Automatic retry, connection reset

3. **Resource Exhaustion**
   - Impact: Service degradation
   - Mitigation: Resource limits, monitoring
   - Recovery: Automatic scaling, manual intervention

## 8. Key Management Details

### 8.1 Key Configuration
The proxy supports two ways to configure encryption keys:

1. **Global Key Configuration**
   - Set via `GCP_KMS_RESOURCE_NAME` environment variable
   - Format: `projects/<project_id>/locations/global/keyRings/<key_ring>/cryptoKeys/<key>`
   - Used as default encryption key for all buckets

2. **Per-Bucket Key Mapping**
   - Set via `GCP_KMS_BUCKET_KEY_MAPPING` environment variable
   - Format: `bucket1:key1,bucket2:key2`
   - Supports wildcard mapping with `*` for global default
   - Example: `"bucket1:projects/project1/locations/global/keyRings/keyring1/cryptoKeys/key1,bucket2/path/to/data:projects/project2/locations/global/keyRings/keyring2/cryptoKeys/key2"`

### 8.2 Key Resolution Logic
1. Global key (`*`) takes highest priority if specified
2. Specific bucket key mapping is checked next
3. If no mapping exists, the bucket passes through unencrypted
4. Keys are validated at startup using a test encryption

### 8.3 Key Metadata Storage
- Encryption key information is stored in GCS object metadata
- Metadata fields:
  - `x-encryption-key`: KMS key resource name used
  - `x-unencrypted-content-length`: Original file size
  - `x-md5Hash`: MD5 hash of unencrypted content
  - `x-proxy-version`: Proxy version for compatibility

### 8.4 Key Usage
- Keys are used for both encryption and decryption operations
- Each operation is performed using Google Cloud KMS
- Tink library provides the cryptographic operations
- AES-256-GCM is used as the underlying encryption algorithm

### 8.5 Key Security
- Keys are never stored locally
- All cryptographic operations are performed in memory
- Key access is managed through Google Cloud IAM
- No key material is logged or exposed in error messages

## 9. Tink Encryption Implementation

### 9.1 Core Components
- Uses Google's Tink library (v1.7.0) for cryptographic operations
- Integrates with Google Cloud KMS for key management
- Implements envelope encryption pattern for secure data handling

### 9.2 Encryption Process
1. **Key URI Construction**
   - Format: `gcp-kms://projects/<project>/locations/<location>/keyRings/<keyring>/cryptoKeys/<key>`
   - Used to identify and access the KMS key

2. **KMS Client Setup**
   - Creates a KMS client using the key URI
   - Registers the KMS client with Tink's registry
   - Establishes secure connection to Google Cloud KMS

3. **Envelope Encryption**
   - Uses Tink's `KMSEnvelopeAEAD2` for envelope encryption
   - Implements AES-256-GCM as the underlying encryption algorithm
   - Provides authenticated encryption with associated data (AEAD)

4. **Data Encryption**
   - Encrypts data in memory
   - Uses empty associated data (AAD) for encryption
   - Maintains encryption metrics for monitoring

### 9.3 Decryption Process
1. **Key Retrieval**
   - Retrieves encryption key ID from object metadata
   - Constructs KMS client for the specific key
   - Validates key access permissions

2. **Data Decryption**
   - Uses the same envelope encryption pattern
   - Decrypts data in memory
   - Maintains decryption metrics for monitoring

### 9.4 Performance Considerations
- Encryption/decryption operations are performed in memory
- Metrics are collected for operation timing
- Supports large file operations (up to 10TB)
- Performance impact on checkpointing:
  - Write: ~60s (encrypted) vs ~50s (unencrypted)
  - Restore: ~40s (encrypted) vs ~33s (unencrypted)

### 9.5 Security Features
- No key material stored locally
- All cryptographic operations performed in memory
- Secure key access through Google Cloud IAM
- Safe logging practices to prevent key exposure
- Uses approved cryptographic algorithms:
  - AES-256-GCM for data encryption
  - SHA-256 for hashing
  - RSA-2048 and ECDSA-256 for key operations

## 10. KMS Implementation Details

### 10.1 KMS Client Setup
- Uses Google Cloud KMS client with Tink integration
- Client initialization with key URI format: `gcp-kms://projects/<project>/locations/<location>/keyRings/<keyring>/cryptoKeys/<key>`
- Supports both global and per-bucket key configurations
- Validates KMS access at startup with test encryption

### 10.2 KMS Client Operations
1. **Client Creation**
   - Uses `gcpkms.NewClientWithOptions` for client initialization
   - Registers KMS client with Tink's registry
   - Creates AEAD (Authenticated Encryption with Associated Data) client

2. **Key Access**
   - Uses Google Cloud IAM for key access control
   - Supports application default credentials
   - Validates key permissions at runtime

3. **Envelope Encryption**
   - Uses `KMSEnvelopeAEAD2` for envelope encryption
   - Implements AES-256-GCM as the underlying algorithm
   - Maintains empty associated data (AAD) for encryption

### 10.3 KMS Integration Features
- Supports large file operations (up to 10TB)
- Implements metrics collection for operation timing
- Provides secure key rotation capabilities
- Maintains compatibility with GCS metadata

### 10.4 KMS Security Measures
- No key material stored locally
- All cryptographic operations performed in memory
- Secure key access through Google Cloud IAM
- Safe logging practices to prevent key exposure

## 11. Operational Procedures

### 11.1 Deployment
1. **Standalone Deployment**
   ```bash
   ./go-gcsproxy -debug=1 \
     -kms_resource_name=projects/PROJECT_ID/locations/global/keyRings/KEYRING/cryptoKeys/KEY \
     -cert_path=/path/to/certs
   ```

2. **Docker Deployment**
   ```bash
   docker run -it \
     -v ${HOME}/.config/gcloud:/path/to/adc \
     -v ${HOME}/certs:/path/to/certs \
     --env-file config.env \
     -p 9080:9080 \
     go-gcsproxy
   ```

3. **Kubernetes Deployment**
   - Sidecar container configuration
   - Service and ingress setup
   - ConfigMap and Secret management

### 11.2 Maintenance
1. **Certificate Management**
   - Certificate rotation procedures
   - CA certificate distribution
   - Certificate validation

2. **Key Rotation**
   - KMS key rotation schedule
   - Key version management
   - Rotation procedures

3. **Logging and Monitoring**
   - Log rotation and retention
   - Metric collection and analysis
   - Alert configuration

### 11.3 Backup and Recovery
1. **Configuration Backup**
   - Environment variables
   - Certificate storage
   - Key mappings

2. **Recovery Procedures**
   - Service restoration
   - Certificate recovery
   - Key recovery

3. **Disaster Recovery**
   - Multi-region deployment
   - Failover procedures
   - Data recovery

## 12. Testing

### 12.1 Test Types
- Functional testing
- Performance testing
- Security testing
- Integration testing

### 12.2 Test Coverage
- Various GCS clients
- Different file sizes
- Multiple encryption scenarios
- Error conditions

## 13. Integration

### 13.1 Client Integration
1. **gsutil Integration**
   ```bash
   export https_proxy=http://127.0.0.1:9080
   export REQUESTS_CA_BUNDLE=/path/to/certs/mitmproxy-ca.pem
   ```

2. **gcloud Integration**
   ```bash
   gcloud config set custom_ca_certs_file /path/to/certs/mitmproxy-ca.pem
   ```

3. **TensorFlow Integration**
   - Custom storage handler
   - Authentication configuration
   - Performance optimization

### 13.2 API Integration
1. **REST API**
   - Endpoint specifications
   - Authentication methods
   - Rate limiting

2. **gRPC API**
   - Service definitions
   - Protocol buffers
   - Streaming support

3. **Third-Party Tools**
   - Compatible tools list
   - Integration procedures
   - Configuration examples

## 14. Glossary

### 14.1 Terms
- **AEAD**: Authenticated Encryption with Associated Data
- **GCS**: Google Cloud Storage
- **KMS**: Key Management Service
- **MITM**: Man-in-the-Middle
- **Tink**: Google's cryptographic library

### 14.2 Acronyms
- **GCP**: Google Cloud Platform
- **IAM**: Identity and Access Management
- **SSL**: Secure Sockets Layer
- **TLS**: Transport Layer Security

## 15. References

1. [Google Cloud KMS Documentation](https://cloud.google.com/kms/docs)
2. [Tink Cryptographic Library](https://github.com/google/tink)
3. [Google Cloud Storage Documentation](https://cloud.google.com/storage/docs)
4. [OpenTelemetry Documentation](https://opentelemetry.io/docs/)

## 16. Version History

### v0.3 (Current)
- Added KMS integration
- Improved error handling
- Enhanced monitoring
- Added Docker support

### v0.2
- Added per-bucket key mapping
- Improved performance
- Enhanced logging

### v0.1
- Initial release
- Basic encryption support
- gsutil integration 