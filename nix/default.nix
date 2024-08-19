{ sources ? import ./sources.nix, system ? builtins.currentSystem, ... }:

import sources.nixpkgs {
  overlays = [
    (_: pkgs: {
      flake-compat = import sources.flake-compat;
      go = pkgs.go_1_21;
      go-ethereum = pkgs.callPackage ./go-ethereum.nix {
        inherit (pkgs.darwin) libobjc;
        inherit (pkgs.darwin.apple_sdk.frameworks) IOKit;
        buildGoModule = pkgs.buildGo121Module;
      };
    }) # update to a version that supports eip-1559
    (import "${sources.poetry2nix}/overlay.nix")
    (import "${sources.gomod2nix}/overlay.nix")
    (pkgs: _:
      import ./scripts.nix {
        inherit pkgs;
        config = {
          ethermint-config = ../scripts/ethermint-devnet.yaml;
          geth-genesis = ../scripts/geth-genesis.json;
          dotenv = builtins.path { name = "dotenv"; path = ../scripts/.env; };
        };
      })
    (_: pkgs: { test-env = pkgs.callPackage ./testenv.nix { }; })
    (_: pkgs: {
      cosmovisor = pkgs.buildGo121Module rec {
        name = "cosmovisor";
        src = sources.cosmos-sdk + "/cosmovisor";
        subPackages = [ "./cmd/cosmovisor" ];
        vendorHash = "sha256-b5WxrM1L2e/J6ZrOKwzmi85YuoRw/bPor20zNIenYS8=";
        doCheck = false;
      };
    })
  ];
  config = { };
  inherit system;
}
