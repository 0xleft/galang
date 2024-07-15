set -e

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit
fi

if ! [ -x "$(command -v wget)" ]; then
  echo "Error: wget is not installed. Please install wget using your package manager."
  exit 1
fi

wget https://github.com/0xleft/gal/releases/latest/download/gal -O /usr/bin/gal
chmod +x /usr/bin/gal
echo "Gal installed in /usr/bin/gal"
echo "Run gal --help to get started."