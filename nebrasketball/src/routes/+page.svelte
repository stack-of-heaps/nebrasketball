<script>
	import { onMount } from 'svelte';
	let messagePromise = getRandomMessage();
	async function getRandomMessage() {
		const response = await fetch('http://localhost:8080/messages/random');
		const data = await response.json();
		return data;
	}

	let message;
	async function getRandomMessage2() {
		const responseObject = await fetch('http://localhost:8080/messages/random').then((r) => r.json())
        message.timestamp = 
	}
	onMount(getRandomMessage2);
</script>

<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://kit.svelte.dev">kit.svelte.dev</a> to read the documentation</p>

{#await messagePromise}
	<p>Getting random message...</p>
{:then message}
	<p>{message.Sender}</p>
	<p>{message.Content}</p>
{:catch error}
	<p>Oops: {error}</p>
{/await}

{#if message}
	<p>Test: {message}</p>
{:else}
	<p>:(</p>
{/if}

<button on:click={() => (messagePromise = getRandomMessage())}> Explore the next planet </button>
