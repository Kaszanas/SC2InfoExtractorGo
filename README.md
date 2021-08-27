[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.5296788.svg)](https://doi.org/10.5281/zenodo.5296788)

# SC2InfoExtractorGo

This tool is meant to allow for quick data extraction from SC2 replay files ".SC2Replay".

## Usage

In order to use this tool please call ```SC2InfoExtractorGo.exe``` and set the choosen flags listed below:

```
  -compression_method int
    	Provide a compression method number, default is 8 'Deflate', other compression methods need to be registered manually in code. (default 8)
  -game_mode int
    	Provide which game mode should be included from the processed files in a format of a binary flag: AllGameModes: 0b1111111111 (default 1023)
  -input string
    	Input directory where .SC2Replay files are held. (default "./DEMOS/Input")
  -integrity_check
    	If the software is supposed to check the hardcoded integrity checks for the provided replays (default true)
  -localize_maps
    	Set to false if You want to keep the original (possibly foreign) map names. (default true)
  -localized_maps_file string
    	Specify a path to localization file containing {'ForeignName': 'EnglishName'} of maps. (default "./operation_files/output.json")
  -log_dir string
    	Provide directory which will hold the logging information. (default "./logs/")
  -log_level int
    	Provide a log level from 1-7. Panic - 1, Fatal - 2, Error - 3, Warn - 4, Info - 5, Debug - 6, Trace - 7 (default 4)
  -number_of_packages int
    	Provide a number of packages to be created and compressed into a zip archive. Please remember that this number needs to be lower than the number of processed files. (default 1)
  -output string
    	Output directory where compressed zip packages will be saved. (default "./DEMOS/Output")
  -perform_anonymization
    	Provide if the tool is supposed to perform the anonymization functions within the processing pipeline. (default true)
  -perform_cleanup
    	Provide if the tool is supposed to perform the cleaning functions within the processing pipeline. (default true)
  -validity_check
    	Provide if the tool is supposed to use hardcoded validity checks and verify if the replay file variables are within 'common sense' ranges. (default true)
  -with_cpu_profiler string
    	Set path to the file where pprof cpu profiler will save its information. If this is empty no profiling is performed.
  -with_multiprocessing
    	Provide if the processing is supposed to be perform with maximum amount of available cores. If set to false, the program will use one core. (default true)
```


## Build from source

Our working solution was built by using ```go build``` command on 64 bit version of Windows 10

## License

This repository is licensed under GNU GPL v3 license. If You would like to acquire a different license please contact me directly.

## Cite Us!

```
@software{Bialecki_2021_5296789,
  author       = {Białecki, Andrzej and
                  Białecki, Piotr},
  title        = {{Kaszanas/SC2InfoExtractorGo: 0.5.0 
                   SC2InfoExtractorGo Release}},
  month        = aug,
  year         = 2021,
  publisher    = {Zenodo},
  version      = {0.5.0},
  doi          = {10.5281/zenodo.5296788},
  url          = {https://doi.org/10.5281/zenodo.5296788}
}
```
