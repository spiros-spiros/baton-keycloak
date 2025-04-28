# Baton Keycloak Connector

**Baton Keycloak Connector** is a plugin that integrates [Keycloak](https://www.keycloak.org/) with [Baton](https://github.com/conductorone/baton), enabling seamless synchronization and provisioning of users and groups.

## 🔧 Features

- **User & Group Synchronization**: Fetches users and groups from Keycloak for Baton to manage.
- **Provisioning Support**: Allows Baton to create, update, and delete groups within Keycloak.
- **Read-Only Mode**: Option to operate in a non-destructive mode, preventing any changes to Keycloak data.
- **Customizable Configuration**: Supports various Keycloak setups through environment variables or command-line flags.

## 🚀 Getting Started

### Prerequisites

- Go 1.23 or higher  
- Access to a running Keycloak instance  
- [Baton CLI](https://github.com/conductorone/baton) installed  

### Installation

Clone the repository:
```git clone https://github.com/spiros-spiros/baton-keycloak.git cd baton-keycloak```

Build the connector: 
```go build -o baton-keycloak```

### Configuration

Set the following environment variables or pass them as command-line flags:

- `KEYCLOAK_URL`: Base URL of your Keycloak instance (e.g., `https://keycloak.example.com`)
- `KEYCLOAK_REALM`: Name of the realm to connect to
- `KEYCLOAK_CLIENT_ID`: Client ID for authentication
- `KEYCLOAK_CLIENT_SECRET`: Client secret for authentication

### Usage

Run the connector:
```./baton-keycloak```

This will start the connector, allowing Baton to interact with your Keycloak instance.

## 🐳 Docker Support

A `Dockerfile` is included for containerized deployments.

Build the Docker image:
```docker build -t baton-keycloak .```

Run the container:
```docker run -e KEYCLOAK_URL=https://keycloak.example.com```
```-e KEYCLOAK_REALM=your-realm```
```-e KEYCLOAK_CLIENT_ID=your-client-id```
```-e KEYCLOAK_CLIENT_SECRET=your-client-secret```
```--provisioner```
```baton-keycloak```

## 📄 License

This project is licensed under the Apache 2 License

## 🤝 Contributing

Contributions are welcome! Please open issues or submit pull requests for any enhancements or bug fixes.

## 📫 Contact

For questions or support, please open an issue on the [GitHub repository](https://github.com/spiros-spiros/baton-keycloak/issues).
