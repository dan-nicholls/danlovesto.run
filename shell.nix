{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
	name = "danlovesto.run Dev Shell";

	buildInputs = [
		pkgs.curl
		pkgs.jq
		pkgs.gnupg
		pkgs.mask
	];

	shellHook = ''
		echo "Loading Strava secrets via gpg..."
		alias strava-env="gpg --decrypt ./strava-secrets.json.gpg | jq -r 'to_entries|map(\"\(.key | ascii_upcase)=\(.value | @sh)\")|.[]' > .env && echo '✅ .env created from GPG'"
		alias smask="mask --maskfile ./docs/strava-api.md"

		echo "danlovesto.run Dev Shell ready"
	'';
}
