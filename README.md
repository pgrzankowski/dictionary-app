# Dictionary App

A GraphQL API built with Go, gqlgen, and GORM for managing translations of Polish words into English. The project uses PostgreSQL as the database and Docker Compose for containerization.

## Prerequisites

- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://golang.org/) (go version go1.23.6)

## Setup & Running Locally

1. **Install Go (go1.23.6)**

2. **Clone the Repository**

   ```sh
   git clone https://github.com/pgrzankowski/dictionary-app.git
   cd dictionary-app
   ```

3. **Create a `.env` File**

   In the root directory, create a file named `.env` with your database and application configuration. For example:

   ```dotenv
    DB_HOST=db
    DB_USER=postgres
    DB_PASS=password
    DB_NAME=dbname
    DB_PORT=5432

    DB_TEST_HOST=localhost
    DB_TEST_USER=admin
    DB_TEST_PASS=pass
    DB_TEST_NAME=dictionary-db-test
    DB_TEST_PORT=5430
   ```

4. **Run the Application**

   Use Docker Compose to build and start the containers:

   ```sh
   docker compose up --build
   ```

   This will:
   - Build the Go application.
   - Start the PostgreSQL container.
   - Run the API server.

5. **Access the GraphQL Playground**

   Open your browser and navigate to [http://localhost:8080](http://localhost:8080) to access the GraphQL Playground and interact with the API.

## Running Tests

To run unit tests execute:

```sh
go test -v -race ./services/
```

This will run all tests since only services are tested.

## Project Structure

- **db/**: Contains database connection logic.
- **graph/**: Contains the GraphQL schema and resolvers.
- **services/**: Contains logic for managing translations.
- **models/**: Contains GORM models for the database tables.
- **.env**: Environment configuration file.

## Additional Information

- **Technology Stack:**
  - **Backend:** Go (go1.23.6), gqlgen, GORM
  - **Database:** PostgreSQL
  - **Containerization:** Docker & Docker Compose

- **Environment Configuration:**  
  Database configuration is managed through the `.env` file.

- **Testing:**  
  Tests are based on exact copy of the main database to provide real value. To run them the docker container with test database must be running.

## Query examples

- **Create translation**
   ```
   mutation {
      createTranslation(
         input: {
            polishWord: "pić"
            englishWord: "drink"
            examples: [{ sentence: "Lubi pić wode." }]
         }
      ) {
         id
         englishWord
         polishWord {
            id
            word
            createdAt
            updatedAt
         }
         examples {
            id
            sentence
            createdAt
            updatedAt
         }
         createdAt
         updatedAt
      }
   }
   ```

- **Remove translation**
   ```
   mutation {
      removeTranslation(id: "1")
   }
   ```

- **Update translation**
   ```
   mutation {
      updateTranslation(
         input: {
            id: "3"
            englishWord: "chug"
         }
      ) {
         id
         englishWord
         polishWord {
            id
            word
            createdAt
            updatedAt
         }
         examples {
            id
            sentence
            createdAt
            updatedAt
         }
         createdAt
         updatedAt
      }
   }
   ```

- **Get all translations**
   ```
   query {
      translations {
         id
         englishWord
         polishWord {
            id
            word
            createdAt
            updatedAt
         }
         examples {
            id
            sentence
            createdAt
            updatedAt
         }
         createdAt
         updatedAt
      }
   }
   ```

- **Get translation by id**
   ```
   query {
      translation(id: "3") {
         id
         englishWord
         polishWord {
            id
            word
            createdAt
            updatedAt
         }
         examples {
            id
            sentence
            createdAt
            updatedAt
         }
         createdAt
         updatedAt
      }
   }
   ```