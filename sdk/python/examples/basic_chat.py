"""
Tortoise Python SDK - Basic Chat Example
"""

import asyncio
from tortoise import TortoiseClient


async def main():
    # Connect to gateway
    async with TortoiseClient(
        gateway_url="http://localhost:8080",
        api_key="your-api-key"
    ) as client:
        # Check health
        health = await client.health_check()
        print(f"Gateway status: {health['status']}")
        
        # Create a new session
        session = await client.create_session(title="My First Chat")
        print(f"Created session: {session.id}")
        
        # Send a chat message
        response = await client.chat(
            message="Hello! What can you do?",
            session_id=session.id
        )
        print(f"AI: {response.content}")
        
        # Continue conversation
        response2 = await client.chat(
            message="Tell me more about yourself.",
            session_id=session.id
        )
        print(f"AI: {response2.content}")


if __name__ == "__main__":
    asyncio.run(main())
