# Death Note Application Backend (Go)

Este repositorio contiene el **backend** de la aplicación Death Note, desarrollado en Go y corriendo sobre PostgreSQL. Está diseñado para ser consumido por un frontend en React a través de una API REST.

---

## 📋 Contenido

* **`• main.go`**: Punto de entrada.
* **`• server/`**: Implementación del servidor, routers y handlers.
* **`• repository/`**: Repositorios para acceso a datos (GORM + PostgreSQL).
* **`• models/`**: Entidades `Person` y `Kill` con conversores a DTO.
* **`• api/`**: DTOs de request/response.
* **`• config/config.json`**: Configuración (puerto, DB, duraciones de kill).
* **`• Dockerfile`, `docker-compose.yml`**: Para contenerización Docker.
* **`• server/server.go`**: Inicialización, migraciones y setup de rutas.
* **`• server/task_queue.go`**: Cola de tareas asincrónicas.
* **`• tests/`** (o integrados en \*\*`repository/`, `server/`): Pruebas unitarias e integración.

---

## ⚙️ Requisitos

* Go 1.24+
* Docker & Docker Compose
* Node.js (para frontend)

---

## 🛠️ Configuración y Ejecución

1. Clona el repositorio:

   ```bash
   git clone https://github.com/tu-usuario/deathnote-backend.git
   cd deathnote-backend
   ```

2. cp .env
   
# Ajusta POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD
  ```bash
  POSTGRES_HOST=localhost
  POSTGRES_DB=deathnote
  POSTGRES_USER=postgres
  POSTGRES_PASSWORD=postgres
  ```

3. Levanta los contenedores Docker:

   ```bash
   docker-compose up --build -d
   ```

4. Verifica que los servicios están corriendo:

   ```bash
   docker ps
   ```

5. El backend estará disponible en `http://localhost:8000`.

---

## 🧪 Pruebas Unitarias y de Integración

### Variables de entorno (según tu shell)

* **PowerShell**:

  ```powershell
  $env:POSTGRES_HOST = "localhost"
  $env:POSTGRES_USER = "postgres"
  $env:POSTGRES_PASSWORD = "postgres"
  $env:POSTGRES_DB = "deathnote"
  ```

* **Bash/Zsh**:

  ```bash
  export POSTGRES_HOST=localhost
  export POSTGRES_USER=postgres
  export POSTGRES_PASSWORD=postgres
  export POSTGRES_DB=deathnote
  ```

### Ejecutar tests

Desde la raíz del proyecto:

```bash
# Pruebas de repositorios
go test ./repository

# Pruebas de servidor y handlers
go test ./server -v

# Todo junto
go test ./...
```

> Con la bandera `-v` verás logs detallados durante la ejecución.

---

## 📡 Endpoints Principales

| Método | Ruta                   | Descripción                                     |
| ------ | ---------------------- | ----------------------------------------------- |
| POST   | `/people`              | Crear persona (multipart: `name`,`age`,`photo`) |
| GET    | `/people`              | Listar todas las personas                       |
| GET    | `/people/{id}`         | Obtener persona por ID                          |
| POST   | `/people/{id}/cause`   | Agregar causa (JSON `{cause}`)                  |
| POST   | `/people/{id}/details` | Agregar detalles (JSON `{details}`)             |
| GET    | `/people/{id}/status`  | Obtener estado actual                           |
| GET    | `/kills`               | Listar kills                                    |
| POST   | `/kills/{id}`          | Crear kill manual (JSON `{description}`)        |

---

## 📖 Frontend (React)

Consulta la [Guía de Integración Frontend](docs/FRONTEND_GUIDE.md) para ver ejemplos de formularios, llamadas HTTP y componentes.

---

## 📝 Licencia

MIT © KevinAPN, UtadLuis msOspina
