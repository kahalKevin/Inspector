import InspectorAgent
from InspectorAgent import agent

# Require to install request and paramiko with:
# pip install requests
# pip install paramiko

agent.startAssertion('127.0.0.1:8001')
# agent.fakeName('77b0cf94-dea6-42db-ba96-33a6362acdd7')

agent.assertMetric("Check int", 10<9, "this check if 10<9")
agent.assertMetric("Check this", True, "this info is additional")

agent.envFreeMemory("103.25.42.117", "root", "root", "free memory of 103.25.42.117")
agent.envFreeMemory("103.25.42.118", "root", "root", "free memory of 103.25.42.118")

agent.finishAssertion()