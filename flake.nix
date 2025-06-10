{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      devShell = pkgs.mkShell {
        name = "Forager";

        buildInputs = with pkgs; [
          air
          bun
          delve
          go
          gofumpt
          gopls
          gotools
          templ
          typescript-language-server
          vscode-langservers-extracted
        ];
      };
    });
}
