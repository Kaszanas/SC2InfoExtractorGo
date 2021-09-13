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
  -game_mode int
    	Provide which game mode should be included from the processed files in a format of a binary flag: AllGameModes: 0b11111111 (default 255)
  -max_procs int
    	Specifies the number of logic cores of a processor that will be used for processing. (default 24)
  -number_of_packages int
    	Provide a number of zip packages to be created and compressed into a zip archive. Please remember that this number needs to be lower than the number of processed files. If set to 0, will ommit the zip packaging and save output .json directly to drive. (default 1)
  -localized_maps_file string
    	Specifies a path to localization file containing {'ForeignName': 'EnglishName'} of maps. If this flag is not set and the default is unavailable, map translation will be ommited. (default "./operation_files/output.json")
  -perform_player_anonymization
    	Specifies if the tool is supposed to perform player anonymization functions within the processing pipeline. If set to true please remember to download and run an anonymization server: https://doi.org/10.5281/zenodo.5138313
  -perform_chat_anonymization
    	Specifies if the chat anonymization should be performed.
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
  -with_cpu_profiler string
    	Set path to the file where pprof cpu profiler will save its information. If this is empty no profiling is performed.
```

### Minimal Example

1. Place ```.SC2Replay``` files in ```./DEMOS/Input```
2. Run ```SC2InfoExtractorGo.exe``` with default flags.
3. Verify the output packages in ```./DEMOS/Output```
4. If The output packages do not contain any processed replays, proceed to verify ```./logs/```.

### Dataset Preparation

If You have a pack of replays with nested directories and You would like to automatically flatten the directory structure, We have published a tool that can be used for that, please see SC2DatasetPreparator: https://doi.org/10.5281/zenodo.5296664

Two scripts contained within that software can:
1. Flatten directory structure of Your collected replays by looking for ```.SC2Replay``` files.
2. Running Python multiprocessing package on multiple directories containing replays by calling them with ```-with-multiprocessing=false``` flag. This allows to have one package per directory created and speed up the processing.

### Anonymization

In order to anonymize the replays please make sure to download and run our open-source implementation of an anonymization server the SC2AnonServerPy: https://doi.org/10.5281/zenodo.5138313

This is required because of the multiprocessing nature of our code that needs to perform synchronization with an existing database unique toons (player IDs) that are mapped to arbitrary incrementing integer.

### Map Translation Support

If the provided file ```output.json``` does not support map names that You require, You will have to either find it online or within the game documents. We are close to publishing a tool that will allow You to download StarCraft maps.

If You already have a set of StarCraft II maps that You would like to create a ```.json``` file to be included within Our software, use Your own solution or SC2MapLocaleExtractor: https://doi.org/10.5281/zenodo.4733264

### Filtering Capabilities

Currently the software supports some game mode filtering capabilities which can be used with ```-game_mode``` flag.
The flag itself is a binary flag where ```0b11111111``` is all game modes which is the default.

Other ways to set the flag:
- ```0b00000001```: 1v1 Ranked Games
- ```0b00000010```: 2v2 Ranked Games
- ```0b00000100```: 3v3 Ranked Games
- ```0b00001000```: 4v4 Ranked Games
- ```0b00010000```: 1v1 Custom Games
- ```0b00100000```: 2v2 Custom Games
- ```0b01000000```: 3v3 Custom Games
- ```0b10000000```: 4v4 Custom Games

## Build from source

Our working solution was built by using ```go build``` command on 64 bit version of Windows 10

## License / Dual Licensing

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
