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
These tokens expire, will need to be refreshed before `goreleaser` can deliver a build to the Github releases area.
- Click [here](https://github.com/settings/tokens?type=beta) while logged in to Github...
- ...or go click-by-click:
  - Log in to the Github web UI.
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

### Build. Package. Sign. Upload.
- Confirm that you have a Github API token.
- Confirm that you know the signing key passphrase.
- Confirm that you're on the `main` branch, and there are no un-committed changes.
  ```shell
  git status
  ```
- Insert the USB stick with the signing key, import the key, and note the key fingerprint. The
  fingerprint should be `4EACB71B2FC20EC8499576BDCB9C922903A66F3F`
  ```shell
  gpg --import "/Volumes/terraform-provider-apstra signing key/terraform-provider-apstra.pgp"
  gpg --list-keys terraform-provider-apstra@juniper.net
  ```
- Put the Github token into the environment.
  ```shell
  # Use `read -s` to get the Github API token into the environment without leaking it into the shell history.
  read -s GITHUB_TOKEN
  # paste the token string and then hit <ctrl>+d
  
  # Use `export` to convert the shell variable into an environment variable.
  export GITHUB_TOKEN
  ```
- Tag the release with a new version number:
  ```shell
  git tag vX.X.X # don't use 'X's, but do use a leading `v`
  ```
- Run `goreleaser` using the recipe in the `Makefile` to publish a draft release to github:
  ```shell
  make release
  ```
- Delete the private key from the keychain:
  ```shell
  gpg --delete-secret-and-public-key terraform-provider-apstra@juniper.net
  ```
- Navigate to the Github Releases [page for this project](https://github.com/Juniper/terraform-provider-apstra/releases).
The new release will be in draft form, not yet visible to the public.
  - Hit the pencil (edit) button for the new release.
  - Press the `Generate release notes` button.
  - Review the release notes, clean up as appropriate.
  - Press the `Publish Release` button.
- Navigate to the [provider page](https://registry.terraform.io/providers/Juniper/apstra) on the Terraform Registry
  - Within a minute or so of publishing the release, the registry should update with the latest version.
