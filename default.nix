{ pkgs }:
pkgs.buildGoModule {
  pname = "dict";
  version = "0.1.0";
  src = ./.;
}
