# Boxtray
A Singbox system tray application based on Clash API and qt6.
## Introduction
Boxtray provides a convenient system tray interface for Singbox management. This project uses the Clash API to communicate with the Singbox core.
## Important Note
This project is based on Singbox's Clash API. You need to configure the Clash API settings yourself to use this application properly.
## Current Features
- System tray interface for Singbox
- Communication via Clash API
## Installation
Currently, no pre-compiled binaries are provided. Please clone the repository and compile it yourself:
```shell
git clone https://github.com/woshikedayaa/boxtray --branch main
cd boxtray
make build
```
## TODO

- Implement cross-platform compilation using GitHub CI/CD
- Optimize initialization logic
- Add support for Meta kernel

## Configuration

see [example.json](example.json)

## Acknowledgments
Thanks to the following libraries:

- [miqt](https://github.com/mappu/miqt) - For providing an easy-to-use and stable Qt binding library.

## Contributing

Contributions are welcome! Feel free to submit issues or pull requests to help improve this project.

## License

This project is licensed under the MIT License