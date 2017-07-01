import InspectorAgent
from InspectorAgent import agent

# Require to install request and paramiko with:
# pip install requests
# pip install paramiko

agent.startAssertion('127.0.0.1:8000')
# agent.fakeName('77b0cf94-dea6-42db-ba96-33a6362acdd7')

agent.assertMetric("check first", True, "this info is additional")
agent.assertMetric("check final", True, "this info is additional")

agent.envFreeMemory("10.99.14.37", "ruma", "ruma1234", "this info is about env")
agent.envFreeMemory("10.99.3.100", "ruma", "ruma1234", "this info is about env")

agent.finishAssertion()