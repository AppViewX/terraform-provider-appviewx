# Certificate Revocation Workflow Status

The `appviewx_revoke_certificate_request_status` resource is used to poll the status of a certificate revocation workflow and view detailed logs and results.

## Process Overview

1. **Workflow Status Polling**:
   - The resource polls the status of a revocation workflow using the `request_id`.
   - Polling is performed at configurable intervals and retry counts.

2. **Status and Logs**:
   - The resource captures the workflow status, status code, summary of all tasks, and detailed logs for any failed tasks.
   - If the workflow fails, the failure reason is extracted from the logs.

3. **State Management**:
   - The resource is read-only. Updates and deletes simply remove the resource from Terraform state.

## Attributes

### Required Attributes

- **`request_id`** (string):  
  The workflow request ID.

### Optional Attributes

- **`retry_count`** (int):  
  Number of times to retry checking workflow status (default: 10).
  Reasonable values - 10 and above

- **`retry_interval`** (int):  
  Seconds to wait between retry attempts (default: 20).
  Reasonable values - 20 and above

## Example Usage

```hcl
resource "appviewx_revoke_certificate_request_status" "revoke_status" {
  request_id    = "<Workflow Request ID>"
  retry_count   = 10
  retry_interval = 20
}
```

## Response

Response after pooling the status of the revoke request

```bash
[CERTIFICATE REVOCATION][SUCCESS] ✅ Operation Result:
{
  "completed_at": "<Timestamp>",
  "operation": "Certificate Revocation",
  "status": "Successful",
  "status_code": 1,
  "workflow_id": "2648"
}
```

## Destroy

To destroy the Certificate details in the Terraform State file, use:

```bash
terraform destroy
```

- This is mainly to ensure that certificates (or any cryptographic material) are not stored in the Terraform state file.
- This feature is crucial for maintaining the security and confidentiality of sensitive cryptographic materials.

---