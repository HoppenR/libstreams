{
  description = "Development Flake for Libstreams";

  inputs = {
    nixpkgs.url = "github:Nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShellNoCC {
          buildInputs = with pkgs; [
            go
            gofumpt
            gopls
          ];

          shellHook = /* bash */ ''
            export LIBSTREAMS_HOME=$(git rev-parse --show-toplevel) || exit
            export XDG_CONFIG_DIRS="$LIBSTREAMS_HOME/.nvim_config:$XDG_CONFIG_DIRS"
          '';
        };
      }
    );
}
