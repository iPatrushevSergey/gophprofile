# Read-only access to GophProfile application secrets.
path "secret/data/gophprofile/*" {
  capabilities = ["read"]
}

path "secret/metadata/gophprofile/*" {
  capabilities = ["read"]
}
