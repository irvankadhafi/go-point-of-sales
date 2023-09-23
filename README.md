# Point Of Sales Service API


## Table of Contents
* [Design Architecture](#design-architecture)
* [Project Structure](#project-structure)

## Design Architecture
The concept of Clean Architecture is used in this service, which means that each component is not dependent on the framework or database used (independent). The service also applies the SOLID and DRY concepts.

Clean Architecture is an architectural pattern that emphasizes separation of concerns in software design.
It is a way of organizing code in a way that makes it easier to maintain, test, and extend over time.
When building a service, using Clean Architecture can have several advantages, including:
- `Separation of concerns`: By separating the business logic of the service from the infrastructure logic, Clean Architecture makes it easier to change or extend the service without affecting other parts of the codebase. This improves maintainability and flexibility.

- `Testability`: Clean Architecture makes it easier to test the service by isolating the business logic from the infrastructure logic. This allows you to write unit tests that focus on the core functionality of the service, without the need to test infrastructure-specific code.

- `Flexibility`: By isolating the business logic from the infrastructure logic, Clean Architecture allows the service to be more easily ported to different infrastructure environments. This improves the reusability and adaptability of the codebase.

- `Dependency Inversion Principle`: Clean Architecture use Dependency Inversion Principle (DIP) principle which makes it easier to change the implementation of a dependency without affecting the code that depends on it.

- `Simplicity`: Clean Architecture promotes a simple and clear codebase, which can make it easier for new team members to understand and maintain.

Overall, using Clean Architecture in building a service can help me to create more `maintainable`, `testable`, and `adaptable code`, which can make it `easier to add new features` and evolve the service over time.


## Project Structure
```bash
.
├── cache/
|   # this package contains code related to managing the cache of your application. 
|   # this could include functions for storing, retrieving, and deleting cache entries.
├── db/migrations
|   # this folder contains SQL files that are used to migrate the database schema.
|   # these files are typically used by a migration tool to update the database schema.
├── internal/
|   #  this folder contains the code of the application that is not intended to be used by external packages
│   ├── config/
│   │   # store configuration files and default values for the application.
│   ├── console/
│   │   # contains script to running server, migrate, create migration.
│   ├── db/
│   │   # contains postgresql and redis init connection.
│   └── delivery/
│   │   # this layer acts as a presenter, providing output to the client.
│   │   # it can use various methods like HTTP REST API, gRPC, GraphQL, etc. In this case, HTTP REST API is used
│   │   └── http/ 
│   │       # this package contains code related to handling HTTP requests and responses.
│   │       # this includes routing, handling requests, and returning responses.
│   │       # this package will mainly act as a presenter, providing output to the client.
│   │   
│   └── helper/
│   │   # this package contains functions that are often used throughout the application.
│   └── model/
│   │   # this layer stores models that will be used by other layers.
│   │   # it can be accessed by all layers
│   └── repository/
│   │   # this layer stores the database and cache handlers.
│   │   # It doesn't contain any business logic and is responsible for determining which datastore to use
│   │   # in this case, RDBMS PostgresSQL is used
│   └── usecase/
│       # this layer contains the business logic for the domain.
│       # it controls which repository to use and performs validation.
│       # it acts as a bridge between the repository and delivery layers
|   
├── utils/
|   # this package contains utility functions that are used throughout the application.
├── config.yml
|   # configuration file to run the server
├── go.mod
├── main.go
├── Makefile
|   # file used by the `make` command
└── ...
```

