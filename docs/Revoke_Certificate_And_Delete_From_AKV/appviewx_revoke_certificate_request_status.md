# Certificate Revocation Workflow Status

The `appviewx_revoke_certificate_request_status` resource is used to poll the status of a certificate revocation workflow and view detailed logs and results.

## Process Overview

1. **Workflow Status Polling**:
   - The resource polls the status of a revocation workflow using the `request_id` returned by `appviewx_revoke_certificate`.
   - Polling is performed at configurable intervals and retry counts.

2. **Status and Logs**:
   - The resource captures the workflow status, status code, summary of all tasks, and detailed logs for any failed tasks.
   - If the workflow fails, the failure reason is extracted from the logs.

3. **State Management**:
   - The resource is read-only. Updates and deletes simply remove the resource from Terraform state.

## Attributes

### Required Attributes

- **`request_id`** (string):  
  The workflow request ID from `appviewx_revoke_certificate.request_id`.

### Optional Attributes

- **`retry_count`** (int):  
  Number of times to retry checking workflow status (default: 10).

- **`retry_interval`** (int):  
  Seconds to wait between retry attempts (default: 20).

### Computed Attributes

- **`status_code`** (int):  
  HTTP status code from the workflow status response.

- **`workflow_name`** (string):  
  Name of the workflow.

- **`workflow_status`** (string):  
  Current status of the workflow (e.g., In Progress, Success, Failed).

- **`workflow_status_code`** (int):  
  Status code of the workflow.

- **`task_summary`** (string):  
  Summary of all task statuses.

- **`failed_task_logs`** (string):  
  Detailed logs of any failed tasks.

- **`failure_reason`** (string):  
  Extracted failure reason from failed task logs.

- **`response_message`** (string):  
  Summary response message from the workflow.

- **`success`** (bool):  
  Whether the workflow completed successfully.

- **`completed`** (bool):  
  Whether the workflow has completed (success or failure).

- **`created_by`** (string):  
  User who created the workflow request.

- **`created_time`** (string):  
  Timestamp when the workflow request was created.

- **`completion_time`** (string):  
  Timestamp when the workflow completed or polling ended.

- **`last_polled_time`** (string):  
  Last time the status was polled.

## Example Usage

```hcl
resource "appviewx_revoke_certificate_request_status" "revoke_status" {
  request_id    = "<Workflow Request ID>"
  retry_count   = 10
  retry_interval = 20
}
```

## Import

To import an existing workflow status into the Terraform state, use:

```bash
terraform import appviewx_revoke_certificate_request_status.revoke_status <request_id>
```
Replace `<request_id>` with the actual workflow request ID.

---