{
  description = "dict";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/release-25.11";
  inputs.util.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, util }: util.lib.eachDefaultSystem (system:
    let pkgs = nixpkgs.legacyPackages.${system}; in
    {
      inherit self nixpkgs;

      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          go
          sqlite
        ];
      };

    });
}

