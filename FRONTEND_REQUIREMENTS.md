# Frontend Requirements

This document describes what the frontend needs to implement to integrate with the new backend features.

## New API Endpoints

### Document Secret Key
- `PUT /api/documents/{id}/secret-key` - Set a secret key on a document (body: `{ "secretKey": "..." }`)
- `DELETE /api/documents/{id}/secret-key` - Remove the secret key from a document

### Document Download (updated)
- `GET /api/documents/{id}?secretKey=...` - Download URL now accepts optional `secretKey` query param
- If a document has a secret key set, the request will fail with 403 unless the correct key is provided

### Document Sharing
- `GET /api/documents/{id}/shares` - List users a document is shared with
- `POST /api/documents/{id}/shares` - Share document with a user (body: `{ "userEmail": "..." }`)
- `DELETE /api/documents/{id}/shares/{shareId}` - Revoke a user share
- `GET /api/documents/{id}/group-shares` - List groups a document is shared with
- `POST /api/documents/{id}/group-shares` - Share document with a group (body: `{ "groupId": "..." }`)
- `DELETE /api/documents/{id}/group-shares/{shareId}` - Revoke a group share

### Document Views
- `GET /api/documents/{id}/views?page=1&pageSize=20` - View history for a document (owner/staff/admin only)

### Groups
- `GET /api/groups` - List groups the user owns or is a member of
- `POST /api/groups` - Create a group (body: `{ "name": "...", "description": "..." }`)
- `GET /api/groups/{id}` - Get group details with member list
- `DELETE /api/groups/{id}` - Delete a group (creator only)
- `POST /api/groups/{id}/members` - Add member to group (body: `{ "userEmail": "..." }`)
- `DELETE /api/groups/{id}/members/{userId}` - Remove member from group

### Document Delete (updated)
- `DELETE /api/documents/{id}` - Now allowed for document owner (previously staff/admin only)

## Updated Response Fields

### DocumentResponse
Two new fields:
- `hasSecretKey` (bool) - Whether the document has a secret key set
- `isOwner` (bool) - Whether the requesting user owns this document

### Document List (updated)
- Clients now see their own documents + documents shared with them (directly or via groups)

## New Pages/Views Needed

### Document Detail/Management Page
- Show document info with `hasSecretKey` and `isOwner` indicators
- If owner: buttons to set/remove secret key, share with users/groups, view history, delete
- If shared user: download button (with secret key prompt if needed)
- Share management section showing current user shares and group shares

### Secret Key Dialog
- Prompt shown when downloading a document that has `hasSecretKey: true`
- Text input for the secret key, submit to download

### Groups Page
- List of groups the user belongs to or created
- Create group form (name + description)
- Group detail view showing members
- Add member form (by email)
- Remove member button (creator only)
- Delete group button (creator only)

### Share Dialog/Panel
- Search for users by email to share a document
- Select a group to share with
- List current shares with revoke buttons

### View History Panel
- Paginated table showing who viewed the document, when, and from what IP
- Only visible to document owner, staff, and admin

## User Flows

### Upload with Secret Key
1. User initiates upload
2. Optionally enters a secret key in the upload form
3. Secret key is sent in the upload request body as `secretKey`

### Download Protected Document
1. User clicks download on a document where `hasSecretKey: true`
2. Frontend prompts for the secret key
3. Key is passed as `?secretKey=...` query parameter on GET request
4. If wrong: 403 error; if correct: download URL returned

### Share a Document
1. Document owner opens share panel
2. Enters user email or selects a group
3. POST to shares or group-shares endpoint
4. Shared user sees document in their document list

### Manage Groups
1. User navigates to Groups page
2. Creates a new group with name/description
3. Adds members by email
4. When sharing documents, can select from their groups
