# Dictionary App

A GraphQL API built with Go, gqlgen, and GORM for managing translations of Polish words into English. The project uses PostgreSQL as the database and Docker Compose for containerization.

## Prerequisites

- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://golang.org/) (go version go1.23.6)

## Setup & Running Locally

1. **Clone the Repository**

   ```sh
   git clone https://github.com/pgrzankowski/dictionary-app.git
   cd dictionary-app
   ```

2. **Create a `.env` File**

   In the root directory, create a file named `.env` with your database and application configuration. For example:

   ```dotenv
    DB_HOST=db
    DB_USER=postgres
    DB_PASS=password
    DB_NAME=dbname
    DB_PORT=5432
   ```

3. **Run the Application**

   Use Docker Compose to build and start the containers:

   ```sh
   docker compose up --build
   ```

   This will:
   - Build the Go application.
   - Start the PostgreSQL container.
   - Run the API server.

4. **Access the GraphQL Playground**

   Open your browser and navigate to [http://localhost:8080](http://localhost:8080) to access the GraphQL Playground and interact with the API.

## Running Tests

To run unit tests, ensure you have Go installed and then execute:

```sh
go test ./...
```

This will run all tests across your project.

## Project Structure

- **db/**: Contains database connection logic.
- **graph/**: Contains the GraphQL schema and resolvers.
- **services/**: Contains business logic for managing translations.
- **models/**: Contains GORM models for your database tables.
- **.env**: Environment configuration file (do not commit sensitive information).

## Additional Information

- **Technology Stack:**
  - **Backend:** Go, gqlgen, GORM
  - **Database:** PostgreSQL
  - **Containerization:** Docker & Docker Compose

- **Environment Configuration:**  
  Database configuration is managed through the `.env` file.

- **Testing:**  
  Unit tests use Go's built-in testing package along with [sqlmock](https://github.com/DATA-DOG/go-sqlmock) for mocking database interactions.
