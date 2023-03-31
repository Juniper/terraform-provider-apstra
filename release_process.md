# Terraform provider key generation/handling/signing protocol

### Initial Key Generation
One time only. This should not be repeated unless the key is compromised or lost.
- Insert a blank USB stick.
- Launch the MacOS disk utility:
  ```shell
  open "/System/Applications/Utilities/Disk Utility.app"
  ```
- In the *Disk Utility* application:
  - click on the disk icon (not a partition) to select it.
  - select `Erase`
    - Name: `terraform-provider-apstra signing key`
    - Format: `Mac OS Extended (Journaled)`
    - Scheme: `Master Boot Record`
    - `[Erase]`
- Generate a new keypair (enter a passphrase when prompted):
  ```shell
  cat << EOF | gpg --batch --gen-key
  Key-Type: 1
  Key-Length: 4096
  Subkey-Type: 1
  Subkey-Length: 4096
  Name-Real: Terraform Provider Apstra developers
  Name-Email: terraform-provider-apstra@juniper.net
  Expire-Date: 0
  EOF
  ```
- Export the private key:
  ```shell
  gpg --output "/Volumes/terraform-provider-apstra signing key/terraform-provider-apstra.pgp" --armor --export-secret-key terraform-provider-apstra@juniper.net
  ```
- Delete the private key from the keychain and eject the USB stick:
  ```shell
  gpg --delete-secret-and-public-key terraform-provider-apstra@juniper.net
  diskutil eject "/Volumes/terraform-provider-apstra signing key"
  ```
  
### Github API Token creation
These tokens expire, will need to be refreshed before `goreleaser` can deliver a build to the github releases area.
- Click [here](https://github.com/settings/tokens?type=beta) while logged in to github...
- ...or go click-by-click:
  - Log in to the github web UI.
  - Click on your avatar at the top right, and then *settings*.
  - Click on *<> Developer settings*.
  - Click on *Personal access tokens* and then *Fine-grained tokens*.
- If an appropriate token has been previously created, but is expired, it can be refreshed:
  - Click on the token.
  - Click on `[Regenerate token]`
  - Select an expiration interval and then click `[Regenerate token]`
  - Copy the token string.
- If an appropriate token does not exist, create one:
  - Click on `[Generate new token]`
  - Token name: `Release terraform-provider-apstra`
  - Resource owner: `Juniper`
  - Repository access
    - *Only select repositories*
    - Select `terraform-provider-apstra`
  - Repository permissions  
    - Contents: Read and Write
    - Metadata: Read-only
  - `[Generate token]`
  - Note the token string. If you're planning to hang onto the token think about how you'll store it securely.
  It's a nuclear bomb, definitely better NOT written to disk.

### Release Signing
- Confirm that you have a github API token and know the signing key passphrase.
- Confirm that you're on the `main` branch, and there are no un-committed changes.
  ```shell
  git status
  ```
- Insert the USB stick with the signing key, import the key, and note the key fingerprint. It
  should match the fingerprint used in the `export` command below.
  ```shell
  gpg --import "/Volumes/terraform-provider-apstra signing key/terraform-provider-apstra.pgp"
  gpg --list-keys terraform-provider-apstra@juniper.net
  ```
- Put required strings into the environment.
  - Use `read -s` to get the github API token into the environment without leaking it into the shell history.
  - Use `export` to load the key fingerprint (not a secret).
  ```shell
  read -s GITHUB_TOKEN
  # paste the token string and then hit <ctrl>+d
  export GITHUB_TOKEN
  export GPG_FINGERPRINT=4EACB71B2FC20EC8499576BDCB9C922903A66F3F 
  ################################################################################
  # Optionally print the GITHUB_TOKEN and GPG_FINGERPRINT to ensure they look okay
  ################################################################################
  # printenv GITHUB_TOKEN
  # printenv GPG_FINGERPRINT
  ```
- Tag the release with the new version number:
  ```shell
  git tag vX.X.X # don't use 'X's
  ```
- Run `goreleaser`:
  ```shell
  goreleaser release --rm-dist
  ```
- Delete the private key from the keychain:
  ```shell
  gpg --delete-secret-and-public-key terraform-provider-apstra@juniper.net
  ```


