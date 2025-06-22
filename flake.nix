{
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        devShell = pkgs.mkShell {
          name = "go";
          buildInputs = with pkgs; [
            air
            delve
            go
            gofumpt
            gopls
            gotools
            sqlc
            tailwindcss_4
            templ
          ];
          PLUGINS_DIR = "./tmp/plugins";
        };
      }
    );
}
