
<h1 align="center">Rmapp</h1>
<p align="center">
  
![ascii.png](/readme-files/ascii.png)
![GitHub Repo stars](https://img.shields.io/github/stars/alewtschuk/rmapp?style=for-the-badge)
![GitHub forks](https://img.shields.io/github/forks/alewtschuk/rmapp?style=for-the-badge)
![License](https://img.shields.io/badge/MIT%20-%20license?style=for-the-badge&label=license)
![GitHub top language](https://img.shields.io/github/languages/top/alewtschuk/rmapp?style=for-the-badge&logo=go&logoColor=white&logoSize=auto&label=%20)
![GitHub Release](https://img.shields.io/github/v/release/alewtschuk/rmapp?style=for-the-badge)

</p>


<p align="center">
Rmapp is a MacOS app removal tool for the command line.
</p>

<p align="center">
It deletes both standard .app bundles and associated files stored elsewhere
in your system, securely, with file size reporting, and default safe trashing. No more drag to trash. No more artifacts.
</p>

**Rmapp build:** ![Build](https://github.com/alewtschuk/rmapp/actions/workflows/build.yml/badge.svg)

**Dependancies:** ![Cobra](https://img.shields.io/badge/passing%20-%20passing?style=flat&logo=github&logoColor=%23969DA4&label=cobra) ![Dsutils](https://github.com/alewtschuk/dsutils/actions/workflows/dsutils.yml/badge.svg) ![pfmt](https://github.com/alewtschuk/pfmt/actions/workflows/pfmt.yml/badge.svg)

**Read the technical breakdown here: [Devlog](https://digitalstatic.substack.com/p/engineering-appcleaner-for-the-terminal)**

## 🚀 Features

- 🗑️ Deletes files safely via trashing through native MacOS APIs
- 💥 Allows for complete unsafe deletion via `--force`
- 📂 Preview the size of and the discovered files via `--peek`
- 💾 Can choose to view files with logical or disk size values
- 📦 Can remove just the bundle via `--bundle`
- 📊 Can check application size via `--size`
- 💻 Built natively in Go for MacOS with Objective-C interop
- 🔐 Works with MacOS system security to safely remove protected files with user approval
- **MORE TO COME !!! 🎉**

## Demo
### Lets get some help
![Help](/readme-files/help.png)

### Peek the files associated with the app
![Peek](/readme-files/peek.png)

### Throw them into the trash
![Peek](/readme-files/trash.png)


## ⬇️ Installation

Rmapp offers a variety of installation options to choose from: 

### 🍺 Homebrew
```bash
  brew tap alewtschuk/formulae
  brew install rmapp
```
### 🔗 Install from source using Go
```bash
  go install github.com/alewtschuk/rmapp@latest
```

## Note
Due to how MacOS configures and protects it's system volume, which includes many of the preinstalled MacOS applications, rmapp will not access or delete any applications within the /System/Applications directory.  

## Contributing

Pull requests are more than welcome! If you find bugs or optimizations that are needed please reach out. For major changes, please open an issue first to discuss what you’d like to change. 

## License
MIT © 2025 Alex Lewtschuk

Made with ❤️ for 👫 around the 🌎
