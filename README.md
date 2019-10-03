# Forvo Pronunciation Downloader

Simple tool for downloading a list of word pronunciations from
[Forvo](https://forvo.com)

## Configuration

Configuration values are read from a file named `forvo.json` in the same
directory as the forvo executable.

- Language codes can be found [here](https://forvo.com/languages-codes/).
- API key requires an account and is available [here](https://api.forvo.com).

## Usage

To specify which words are downloaded, enter one word per line in `forvo.txt` in
the same directory as the tool executable.

After configuration and words are set, just run the program. Log messages with
information and errors are logged to `forvo.log`.
