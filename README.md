## Setup Instructions

### Backend
1. Navigate to the `backend` directory:
    ```sh
    cd backend
    ```
2. Install dependencies:
    ```sh
    go mod tidy
    ```
3. Run the backend server:
    ```sh
    go run main.go
    ```

### Frontend
1. Navigate to the `frontend` directory:
    ```sh
    cd frontend
    ```
2. Install dependencies:
    ```sh
    npm install
    ```
3. Run the frontend application:
    ```sh
    npm start
    ```

## API Endpoints

### GET /cache/{key}
- Retrieve the value for a given key from the cache.

### SET /cache
- Add a new key/value pair to the cache with an optional expiration time.

### DELETE /cache/{key}
- Delete a key from the cache.

## WebSocket (Optional)
- Reflects all the current key-values pairs with their expiration time dynamically in the UI.

## Contact
For any queries, please contact [suhasbdvt6@gmail.com].
