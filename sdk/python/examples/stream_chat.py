"""
Tortoise Python SDK - Streaming Chat Example
"""

import asyncio
from tortoise import TortoiseClient


async def main():
    async with TortoiseClient(
        gateway_url="http://localhost:8080",
        api_key="your-api-key"
    ) as client:
        print("Starting streaming chat...\n")
        
        # Stream response
        full_response = ""
        async for chunk in client.chat_stream(
            message="Write a short poem about artificial intelligence.",
            temperature=0.8
        ):
            print(chunk, end="", flush=True)
            full_response += chunk
        
        print("\n\n[Streaming complete]")


if __name__ == "__main__":
    asyncio.run(main())
