const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({ headless: false });
  const page = await browser.newPage();
  
  // Capture console errors
  page.on('console', msg => {
    if (msg.type() === 'error') {
      console.log('BROWSER ERROR:', msg.text());
    }
  });
  
  // Capture network errors
  page.on('response', response => {
    if (response.status() >= 400) {
      console.log('HTTP ERROR:', response.status(), response.url());
    }
  });
  
  try {
    // Navigate to login page
    await page.goto('http://localhost:3000/login');
    await page.waitForSelector('input[name="username"]', { timeout: 10000 });
    
    console.log('Login page loaded');
    
    // Take screenshot to debug
    await page.screenshot({ path: 'login-page.png' });
    console.log('Screenshot saved as login-page.png');
    
    // Find and click any button with text "登录" or "Login"
    const loginButton = await page.$('button');
    if (loginButton) {
      await page.type('input[name="username"]', 'root');
      await page.type('input[name="password"]', 'hWCEW@8##47$RGAK');
      await loginButton.click();
    } else {
      console.log('No login button found');
      return;
    }
    
    // Wait for redirect to home page
    await page.waitForNavigation();
    console.log('Logged in successfully');
    
    // Navigate to channel page
    await page.goto('http://localhost:3000/channel');
    await page.waitForSelector('.ui.table', { timeout: 10000 });
    
    console.log('Channel page loaded');
    
    // Just take a screenshot and check logs manually
    await page.screenshot({ path: 'channel-page.png' });
    console.log('Channel page screenshot saved');
    
    // Try to find any buttons
    const allButtons = await page.$$('button');
    console.log(`Found ${allButtons.length} total buttons`);
    
    // Look for test button by searching button text
    for (let i = 0; i < allButtons.length; i++) {
      const text = await page.evaluate(el => el.textContent, allButtons[i]);
      if (text.includes('测试') || text.toLowerCase().includes('test')) {
        console.log(`Found test button: "${text}"`);
        await allButtons[i].click();
        
        // Wait and capture any errors
        await new Promise(resolve => setTimeout(resolve, 5000));
        
        // Check console errors
        const logs = await page.evaluate(() => {
          return window.console._logs || [];
        });
        
        // Get any error messages
        const messages = await page.$$eval('.ui.message', msgs => 
          msgs.map(m => ({type: m.className, text: m.textContent}))
        );
        
        console.log('Messages on page:', messages);
        console.log('Clicked test button');
        break;
      }
    }
    
  } catch (error) {
    console.error('Error:', error);
  } finally {
    await browser.close();
  }
})();