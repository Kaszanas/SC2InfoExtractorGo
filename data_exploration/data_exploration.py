import json


replay_file = "./data_exploration/sample_data_1.json"

loaded_data = json.load(open(replay_file, encoding='utf-8'))

def return_unique(dict_key:str):
    unique_events = {}

    for event in loaded_data[dict_key]:
        if event["evtTypeName"] in unique_events:
            unique_events[event["evtTypeName"]] += 1
        else:
            unique_events[event["evtTypeName"]] = 1

    return unique_events

def show_unique_keys(dict_key:str):
    unique_keys = set()

    for event in loaded_data[dict_key]:
        for key in event.keys():
            unique_keys.add(key)

    return unique_keys

# TODO: Write a function that specifies in which event Type a key resides so that it is easier to describe them.


try:
    unique_game_events = return_unique("gameEvents")
except:
    pass

try:
    unique_message_events = return_unique("messageEvents")
except:
    pass

try:
    unique_tracker_events = return_unique("trackerEvents")
except:
    pass

print(unique_game_events)
# print(unique_message_events)
print(unique_tracker_events)


unique_game_events_keys = show_unique_keys("trackerEvents")
print(unique_game_events_keys)

