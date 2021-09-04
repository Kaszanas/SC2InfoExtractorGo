[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.5296788.svg)](https://doi.org/10.5281/zenodo.5296788)

# SC2InfoExtractorGo

This tool is meant to allow for quick data extraction from SC2 replay files ".SC2Replay".

## Usage

In order to use this tool please call ```SC2InfoExtractorGo.exe``` and set the choosen flags listed below:

```
  -input string
    	Input directory where .SC2Replay files are held. (default "./DEMOS/Input")
  -output string
    	Output directory where compressed zip packages will be saved. (default "./DEMOS/Output")
  -number_of_packages int
    	Provide a number of zip packages to be created and compressed into a zip archive. Please remember that this number needs to be lower than the number of processed files. (default 1)
  -perform_anonymization
    	Provide if the tool is supposed to perform the anonymization functions within the processing pipeline. If set to true please remember to download and run an anonymization server. https://doi.org/10.5281/zenodo.5138313
  -perform_cleanup
    	Provide if the tool is supposed to perform the cleaning functions within the processing pipeline.
  -perform_integrity_checks
    	If the software is supposed to check the hardcoded integrity checks for the provided replays
  -perform_validity_checks
    	Provide if the tool is supposed to use hardcoded validity checks and verify if the replay file variables are within 'common sense' ranges.
  -log_dir string
    	Specifies directory which will hold the logging information. (default "./logs/")
  -log_level int
    	Specifies a log level from 1-7. Panic - 1, Fatal - 2, Error - 3, Warn - 4, Info - 5, Debug - 6, Trace - 7 (default 4)
  -game_mode int
    	Provide which game mode should be included from the processed files in a format of a binary flag: AllGameModes: 0b1111111111 (default 1023)
  -localized_maps_file string
    	Specifies a path to localization file containing {'ForeignName': 'EnglishName'} of maps. (default "./operation_files/output.json")
  -with_cpu_profiler string
    	Set path to the file where pprof cpu profiler will save its information. If this is empty no profiling is performed.
  -with_multiprocessing
    	Specifies if the processing is supposed to be perform with maximum amount of available cores. If set to false, the program will use one core.
```

### Minimal Example

1. Place ```.SC2Replay``` files in ```./DEMOS/Input```
2. Run ```SC2InfoExtractorGo.exe``` with default flags.
3. Verify the output packages in ```./DEMOS/Output```
4. If The output packages do not contain any processed replays, proceed to verify ```./logs/```.

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
