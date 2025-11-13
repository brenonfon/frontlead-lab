from sdk.action import Action
from shared.tracker.api import RemoteLLMTrackerApi
from loguru import logger
import aiohttp
import json

class action(Action):
  async def run(self, tracker_api: RemoteLLMTrackerApi):
    """ Start coding here """
    try:
      # Example API endpoint - replace with your actual API URL
      api_url = "https://neural-pregnantly-rebeca.ngrok-free.dev/leads?userId=breno.rios@nfon.com"
      
      async with aiohttp.ClientSession() as session:
        async with session.get(api_url) as response:
          if response.status == 200:
            data = await response.json()
            logger.info(f"API response: {data}")
            return data
          else:
            logger.error(f"API call failed with status: {response.status}")
            return {"isExistingLead": False}
            
    except Exception as e:
      logger.error(f"Error making API call: {str(e)}")
      return {"isExistingLead": False}