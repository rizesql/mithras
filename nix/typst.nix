{ pkgs, ... }:
{
  devShell = pkgs.mkShell {
    packages = with pkgs; [
      typst
      tinymist
      gyre-fonts

      just
    ];

    shellHook = ''
      export TYPST_FONT_PATHS="${pkgs.gyre-fonts}/share/fonts"
      echo "Typst: $(typst --version)"
      echo "tinymist: $(tinymist --version)"
      echo "Fonts available at: $TYPST_FONT_PATHS"
    '';
  };
}
