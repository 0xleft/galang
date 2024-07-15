# GenAlpha lang 🤓

A programming language featuring simple AST generation and interpreter

Further design schematics can be found in [docs](docs/README.md) folder.

## Install

### Windows

Download installer from [here](https://github.com/0xleft/gal/releases/latest/download/windows_installer.exe)

```powershell
gal.exe --help
```

### Linux

```bash
curl https://raw.githubusercontent.com/0xleft/gal/main/install.sh | sudo bash
gal --help
```

## Hello world ❤️

```gal
lowkey main{}
    fire std.println("hello world")
end

` or could also be written in one line
lowkey main{};fire std.println("hello world");end
```

## Build from source

```bash
git clone https://github.com/0xleft/gal
cd gal
go build
```

## License

This project and all projects using this programming language must be licensed under the skibidi license such as one featured in this repository.