{
  description = "SC2InfoExtractorGo: extracts data from StarCraft II .SC2Replay files";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "SC2InfoExtractorGo";
          version = "0.1.0";

          src = ./.;

          vendorHash = "sha256-AHgAIq6/wERy0zq634AorOkQQIoZgOVsRXrMIzZXzK8=";

          # The E2E test suite needs fixture data (`make fetch_test_fixtures`)
          # and network access, neither available in the sandboxed build.
          # It's already exercised separately in CI (e2e_tests.yml).
          doCheck = false;

          meta = {
            description = "Extracts data from StarCraft II .SC2Replay files into JSON";
            homepage = "https://github.com/Kaszanas/SC2InfoExtractorGo";
            license = pkgs.lib.licenses.gpl3Only;
            mainProgram = "SC2InfoExtractorGo";
          };
        };

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.golangci-lint
          ];
        };
      });
}
