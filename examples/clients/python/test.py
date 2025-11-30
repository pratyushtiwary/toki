import time
from client import TokiClient, TokiEvent

client = TokiClient()

@client.event_handler
async def handle_event(event_data: TokiEvent):
    if event_data == TokiEvent.AUTH_REFRESHED:
        print("Auth refreshed, refreshing client...")


for i in range(10):

    print(i + 1)

    time.sleep(1)
