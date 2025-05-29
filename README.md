
# Rmapp
<p align="center">
  
![ascii.png](/readme-files/ascii.png)
![GitHub Repo stars](https://img.shields.io/github/stars/alewtschuk/rmapp?style=for-the-badge)
![GitHub forks](https://img.shields.io/github/forks/alewtschuk/rmapp?style=for-the-badge)
![License](https://img.shields.io/badge/MIT%20-%20license?style=for-the-badge&label=license)
![GitHub top language](https://img.shields.io/github/languages/top/alewtschuk/rmapp?style=for-the-badge&logo=go&logoColor=white&logoSize=auto&label=%20)
![GitHub Release](https://img.shields.io/github/v/release/alewtschuk/rmapp?style=for-the-badge)


Rmapp is a MacOS app removal tool for the command line.

It deletes both standard .app bundles and associated files stored elsewhere
in your system, securely, with file size reporting, and default safe trashing. No more drag to trash. No more artifacts.

</p>

**Rmapp build:**\
![Build](https://github.com/alewtschuk/rmapp/actions/workflows/build.yml/badge.svg)

**Dependancies:**\
![Cobra](https://img.shields.io/badge/passing%20-%20passing?style=flat&logo=github&logoColor=%23969DA4&label=cobra)
![Dsutils](https://github.com/alewtschuk/dsutils/actions/workflows/dsutils.yml/badge.svg)
![pfmt](https://github.com/alewtschuk/pfmt/actions/workflows/pfmt.yml/badge.svg)



## 🚀 Features

- 🗑️ Deletes files safely via trashing through native MacOS APIs
- 💥 Allows for complete unsafe deletion via `--force`
- 📂 Preview the size of and the discovered files via `--peek`
- 💾 Can choose to view files with logical or disk size values
- 💻 Built natively in Go for MacOS with Objective-C interop
- 🔐 Works with MacOS System Integrity Protection(SIP) to safely remove protected files with user approval
- **MORE TO COME !!! 🎉**


## ⬇️ Installation

Rmapp offers a variety of installation options to choose from: 

### 🍺 Homebrew
```bash
  brew tap alewtschuk/formulae
  brew install rmapp
```
### 🔗 Install from source using Go
```bash
  git clone https://github.com/alewtschuk/rmapp.git
  cd rmapp
  go install
```
