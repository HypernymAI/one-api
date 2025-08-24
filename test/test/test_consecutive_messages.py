import asyncio
from openai import AsyncOpenAI

async def test():
    client = AsyncOpenAI(
        base_url='http://localhost:3000/v1',
        api_key='sk-BoroIrY7uCruh3IR48Cb5b151f8943E88bF86e3d91Ee2bB2'
    )

    print("Testing with consecutive user messages...")
    
    try:
        response = await client.chat.completions.create(
            model='gpt-4o',
            messages=[
                {"role": "system", "content": "You are a helpful assistant."},
                {"role": "user", "content": "First user message"},
                {"role": "user", "content": "Second user message (consecutive)"}
            ],
            temperature=0.5
        )
        print(f'Success! Response: {response.choices[0].message.content[:50]}...')
    except Exception as e:
        print(f'Error: {type(e).__name__}: {e}')
        if hasattr(e, 'response'):
            print(f'Response status: {getattr(e.response, "status_code", "N/A")}')
            print(f'Response body: {getattr(e.response, "text", "N/A")}')

asyncio.run(test())