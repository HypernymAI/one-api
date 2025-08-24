import asyncio
from openai import AsyncOpenAI
import os
from dotenv import load_dotenv

load_dotenv()

async def test():
    client = AsyncOpenAI(
        base_url='http://localhost:3000/v1',
        api_key=os.getenv('ONE_API_API_KEY', 'sk-BoroIrY7uCruh3IR48Cb5b151f8943E88bF86e3d91Ee2bB2')
    )

    print(f'Client base_url: {client.base_url}')
    print(f'Client api_key: {client.api_key[:20]}...')
    print(f'Testing simple completion...')

    try:
        response = await client.chat.completions.create(
            model='gpt-4o',
            messages=[
                {"role": "system", "content": "You are an expert at finding the simplest hypernyms & synopsis possible."},
                {"role": "user", "content": "Say hello"}
            ],
            temperature=0.5
        )
        print(f'Success! Response: {response.choices[0].message.content}')
        print(f'Response model: {response.model}')
        print(f'Response ID: {response.id}')
    except Exception as e:
        print(f'Error: {type(e).__name__}: {e}')
        if hasattr(e, 'response'):
            print(f'Response status: {e.response.status_code if hasattr(e.response, "status_code") else "N/A"}')
            print(f'Response body: {e.response.text if hasattr(e.response, "text") else "N/A"}')

asyncio.run(test())