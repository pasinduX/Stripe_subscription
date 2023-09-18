/* Handle button clicks */
document.getElementById('basic-plan-btn').addEventListener('click', async () => {
  const priceId = document.querySelector('#basicPrice').value; // Extract the priceId
  if (!priceId) {
    console.error('Price ID is empty'); // Log an error if the priceId is empty
    return;
  }

  const { error } = await createCheckoutSession(priceId); // Pass the extracted priceId
  if (error) {
    displayError(error);
  }
});

document.getElementById('pro-plan-btn').addEventListener('click', async () => {
  const priceId = document.querySelector('#proPrice').value; // Extract the priceId
  if (!priceId) {
    console.error('Price ID is empty'); // Log an error if the priceId is empty
    return;
  }

  const { error } = await createCheckoutSession(priceId); // Pass the extracted priceId
  if (error) {
    displayError(error);
  }
});

/* Function to create a checkout session */
async function createCheckoutSession(priceId) {
  try {
    const response = await fetch('/create-checkout-session', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: `priceId=${encodeURIComponent(priceId)}`, // Send the priceId as a parameter
    });

    if (!response.ok) {
      throw new Error('Failed to create a checkout session');
    }

    const session = await response.json();
    return { session };
  } catch (error) {
    return { error: { message: error.message } };
  }
}

/* Function to display an error message */
function displayError(error) {
  const errorMessage = document.getElementById('error-message');
  errorMessage.textContent = error.message;
  errorMessage.style.display = 'block';
}

/* Fetch prices and update the form */
fetch('/config')
  .then(r => r.json())
  .then(({ basicPrice, proPrice }) => {
    const basicPriceInput = document.querySelector('#basicPrice');
    basicPriceInput.value = basicPrice;
    const proPriceInput = document.querySelector('#proPrice');
    proPriceInput.value = proPrice;
  });
