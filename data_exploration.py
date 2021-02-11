import json


replay_file = "./sample_data.json"

loaded_data = json.load(open(replay_file, encoding='utf-8'))

def return_unique(dict_key:str):
    unique_events = {}

    for event in loaded_data[dict_key]:
        if event["evtTypeName"] in unique_events:
            unique_events[event["evtTypeName"]] += 1
        else:
            unique_events[event["evtTypeName"]] = 1

    return unique_events

print(return_unique("gameEventStrings"))

print(return_unique("messageEventsStrings"))

print(return_unique("trackerEventStrings"))