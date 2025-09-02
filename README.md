# consul-acl-sync

Synchronize Consul ACL policies and tokens from YAML configuration files.

## Overview

`consul-acl-sync` reads ACL configuration from YAML files and applies them to Consul using `plan` and `apply` commands similar to [Terraform](https://developer.hashicorp.com/terraform).

## Installation

```bash
$ go install github.com/zinrai/consul-acl-sync@latest
```

## Quick Start

### 1. Start Consul with ACL enabled

```bash
$ consul agent -dev -hcl 'acl = { enabled = true, default_policy = "allow" }'
```

### 2. Bootstrap ACL system

```bash
$ consul acl bootstrap > bootstrap.json
$ export CONSUL_HTTP_TOKEN=$(cat bootstrap.json | grep SecretID | cut -d'"' -f4)
```

### 3. Create configuration

See `example.yaml` for a complete example.

### 4. Apply configuration

Show what will be changed

```bash
$ consul-acl-sync plan -config example.yaml
```

Apply changes

```bash
$ consul-acl-sync apply -config example.yaml
```

Skip confirmation

```bash
$ consul-acl-sync apply -config example.yaml -auto-approve
```

## Configuration Format

### Policies

- `name`: Policy name (required, unique)
- `description`: Policy description (optional)
- `rules`: HCL format rules (required)

### Tokens

- `description`: Token description (required, unique identifier)
- `policies`: List of policy names (required)

## Environment Variables

- `CONSUL_HTTP_ADDR`: Consul server address (default: `http://localhost:8500`)
- `CONSUL_HTTP_TOKEN`: Consul ACL management token (required)

## How It Works

1. **Read Configuration**: Load policies and tokens from YAML file
2. **Fetch Current State**: Query Consul for existing resources (only those defined in YAML)
3. **Calculate Diff**: Compare current and desired state
4. **Apply Changes**: Create new resources or update existing ones

### Design Principles

- **No Deletion**: Resources are never deleted, only created or updated
- **YAML-Driven**: Only manages resources explicitly defined in YAML
- **Ignore System Resources**: Never touches built-in policies or system tokens
- **Idempotent**: Running the same configuration multiple times produces consistent results

### Limitations

- Only supports policies and tokens (no roles or auth methods)
- Tokens are identified by description field (must be unique)
- Policies are identified by name (must be unique)
- No resource deletion (by design for safety)

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
