

# CnkiPatentSpiderGo

Step 3 of the CNKI Patent Spider (truly distributed). For the previous two steps, see [CnkiPatentSpider](https://github.com/aFlyBird0/CnkiPatentSpider).

It can run on different machines (no additional configuration required, tasks are automatically distributed), and can also be run multiple times simultaneously on the same machine. You can stop it at any time.

## Usage

Locate the binary file in the `bin` directory that matches your operating system and chip architecture, and run `./binary_filename run`.

For example, for macOS with an M1 chip architecture, run `./spider_darwin_arm64 run`.

Run the program in the background and do not close the terminal. Alternatively, you can use commands like `nohup` to run it in the background.

There are additional parameters you can configure. Run `./binary_filename run --help` to view them. Under normal circumstances, no configuration is needed.

For example, to set the minimum interval between requests to 2 seconds and the concurrency to 2: `./binary_filename run --min=2 -c=2`.

## Notes

To run locally, rename `/db/dsn_example.txt` to `/db/dsn.txt` and update the database connection information inside it.

## Cross-Compilation

Run `/build.sh` to automatically compile binaries for various operating systems and chip architectures. The generated binary files will be stored in the `/知网专利爬虫/bin` directory.
