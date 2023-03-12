# This example creates an Agent Profile with a couple of options.
# Note that it is not possible to set device credentials with this
# resource because (a) the credentials cannot be retrieved, so
# terraform cannot manage the full lifecycle and (b) keeping
# secrets in the state file probably isn't great.
resource "apstra_agent_profile" "profile_with_options" {
  name = "spine switches"
  platform = "eos"
  open_options = {
    foo = "bar"
    baz = "not bar"
  }
}