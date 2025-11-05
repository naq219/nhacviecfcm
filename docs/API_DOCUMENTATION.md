# RemiAq API Documentation

## 1. Introduction

This document provides a detailed description of the RemiAq API. It is intended for client-side developers and AI assistants who need to interact with the RemiAq backend.

The API provides functionalities for managing reminders, checking system status, and executing raw SQL queries for administrative purposes.

**Base URL**: `http://localhost:8090`

## 2. Authentication

Most endpoints are protected and require authentication. The API uses PocketBase's built-in token-based authentication.

To authenticate, you need to obtain a token from the PocketBase authentication endpoint (e.g., `/api/collections/users/auth-with-password`) and include it in the `Authorization` header for subsequent requests.

**Header Format**:
```
Authorization: Bearer YOUR_AUTH_TOKEN
```

Endpoints that require authentication are marked with `(Secure)`.

## 3. Endpoints

### 3.1. Health Check

- **Endpoint**: `GET /hello`
- **Description**: A simple endpoint to check if the API server is running.
- **Authentication**: None
- **Success Response**:
  - **Code**: `200 OK`
  - **Content-Type**: `text/plain`
```
  - **Body**:
    ```text
    RemiAq API is running!
    ```
```
### 3.2. System Status

- **Endpoint**: `GET /api/system_status`
- **Description**: Retrieves the current status of the system, including worker status and last error.
- **Authentication**: None
- **Success Response**:
  - **Code**: `200 OK`
  - **Content-Type**: `application/json`
  - **Body**: A `SystemStatus` object.
    ```json
    {
        "success":  true,
        "data":  {
                     "mid":  1,
                     "worker_enabled":  false,
                     "last_error":  "",
                     "updated":  "2025-11-04T15:55:20.079Z"
                 }
    }
    ```

---

- **Endpoint**: `PUT /api/system_status`
- **Description**: Updates the system status. This is typically used for enabling/disabling the background worker or clearing errors.
- **Authentication**: Recommended to be admin-only.
- **Request Body**:
  ```json
  {
    "worker_enabled": true,
    "last_error": "Optional error message"
  }
  ```
- **Success Response**:
  - **Code**: `200 OK`
  - **Content-Type**: `application/json`
  - **Body**: The updated `SystemStatus` object.
    ```json
    {
        "success":  true,
        "message":  "System status updated",
        "data":  {
                     "mid":  1,
                     "worker_enabled":  true,
                     "last_error":  "",
                     "updated":  "2025-11-05T08:21:36.1001069Z"
                 }
    }
    ```

### 3.3. Raw SQL Queries (Admin Only)

These endpoints are for administrative purposes and should be protected with strict access control. They allow executing raw SQL queries against the database.

**Request Body Format** (for POST/PUT):
```json
{
  "query": "SELECT * FROM table_name;"
}
```
**Request Query Parameter** (for GET/DELETE): `?q=SELECT...`

- **Endpoint**: `POST /api/rquery` (or `GET`)
  - **Action**: Executes a `SELECT` query.
- **Endpoint**: `POST /api/rinsert` (or `GET`)
  - **Action**: Executes an `INSERT` query.
- **Endpoint**: `PUT /api/rupdate` (or `GET`)
  - **Action**: Executes an `UPDATE` query.
- **Endpoint**: `DELETE /api/rdelete` (or `GET`)
  - **Action**: Executes a `DELETE` query.

### 3.4. Reminders (Secure)

All reminder endpoints require authentication.

- **Endpoint**: `POST /api/reminders`
  - **Action**: Creates a new reminder.
  - **Request Body**: A `Reminder` object.
  - **Response**: The created `Reminder` object.

- **Endpoint**: `GET /api/reminders/{id}`
  - **Action**: Retrieves a specific reminder by its ID.
  - **Response**: A `Reminder` object.

- **Endpoint**: `PUT /api/reminders/{id}`
  - **Action**: Updates a reminder.
  - **Request Body**: A `Reminder` object.
  - **Response**: The updated `Reminder` object.

- **Endpoint**: `DELETE /api/reminders/{id}`
  - **Action**: Deletes a reminder.
  - **Response**: A success message.

- **Endpoint**: `GET /api/collections/reminders/records`
  - **Action**: Retrieves all reminders for the authenticated user. The API supports filtering, sorting, and pagination via query parameters. See PocketBase API documentation for details.
  - **Response**: A paginated list of `Reminder` objects.

- **Endpoint**: `POST /api/reminders/{id}/snooze`
  - **Action**: Snoozes a reminder for a specified duration.
  - **Request Body**:
    ```json
    {
      "duration": 3600 // in seconds
    }
    ```
  - **Response**: A success message.

- **Endpoint**: `POST /api/reminders/{id}/complete`
  - **Action**: Marks a reminder as complete.
  - **Response**: A success message.




## 4. Data Models

### Reminder
```json
{
  "id": "string (record_id)",
  "user_id": "string (user_record_id)",
  "title": "string",
  "content": "string",
  "type": "string (one_time, daily, weekly, monthly, yearly, custom)",
  "cron_expression": "string (e.g., '0 9 * * 1-5')",
  "due_date": "string (ISO 8601 format, e.g., '2024-01-15T09:00:00Z')",
  "is_lunar": false,
  "status": "string (pending, snoozed, completed, error)",
  "last_error": "string",
  "created": "string (ISO 8601 format)",
  "updated": "string (ISO 8601 format)"
}
```

### SystemStatus
```json
{
  "id": "string (always '1')",
  "worker_enabled": true,
  "last_run_time": "string (ISO 8601 format)",
  "last_error": "string",
  "created": "string (ISO 8601 format)",
  "updated": "string (ISO 8601 format)"
}
```

## 5. Error Responses

When an API call fails, the server responds with a standard JSON error format.

- **Code**: `4xx` or `5xx`
- **Content-Type**: `application/json`
- **Body**:
  ```json
  {
    "status": "error",
    "message": "A descriptive error message",
    "error": "Detailed error information (optional, often omitted in production)"
  }
  ```