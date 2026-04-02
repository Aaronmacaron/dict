# dict

dict is a CLI app to look up words in a [dict.cc](https://www.dict.cc/) dictionary.
It works fully offline by accessing local dict.cc dictionary databases.
dict supports all languages that are supported by dict.cc.

[![asciicast](https://asciinema.org/a/ReojzxAnX4hnlSNu.svg)](https://asciinema.org/a/ReojzxAnX4hnlSNu)

## Getting Started

After you've [installed](#Installation) dict you need to register a dictionary file that dict will use to look up words.
For legal reasons you have to manually download the dictionary file.
I'm not sure if I'm allowed to describe how get ahold of a dictionary file, but it's not very difficult.
If you savvy enough to want to use a dictionary from you terminal you can probably find out how to do it.

Once you have a dictionary file you can it register with dict:

```shell
dict --register path/to/you/dictionary-file.db
```

Then you can set it as the default dictionary:

```shell
dict --default dictionary-file.db
```

You can check all registered dictionaries like that:

```shell
dict --list
  de-en.db (German-English)
  de-it.db (German-Italian)
* en-es.db (English-Spanish)
  en-fr.db (English-French)
```

The asterisk in the output indicates what dictionary is currently set as the default.

Now you're ready to look up words:

```shell
dict <word>
```

You can use the `-a` or `--all` flag to show all results instead only the most relevant ones.
You can use the `-d` or `--dict` flag to temporarily use a dictionary other than the default.


## Installation

There are different ways to install dict.
The installation methods are confirmed to work on macOS and Linux, and manual as well as local compilation should work on Windows as well, but I didn't test it.

### Homebrew

Homebrew is the preferred installation method on macOS:

```shell
brew install --cask Aaronmacaron/homebrew-tap/dict
xattr -d com.apple.quarantine $(which dict)
```

The second command is required because I didn't officially sign the binaries by going through the Apple signing process.
If you use homebrew on Linux you can skip this step.

### Install script (Linux and MacOS)

The installation script installs dict at `/usr/local/bin`.

```shell
curl https://raw.githubusercontent.com/Aaronmacaron/dict/master/install.sh | sh
```

### Manual

You can manually download the latest release from the [GitHub release page](https://github.com/Aaronmacaron/dict/releases) and then run the binary or move it to a dir where's it in the $PATH.


### Compile locally

You can install dict by compiling it locally:

```shell
go install github.com/Aaronmacaron/dict/cmd/dict@latest
```