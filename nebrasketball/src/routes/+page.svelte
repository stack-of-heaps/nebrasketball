<script lang>
	import { onMount } from 'svelte';

	let messagePromise = getRandomMessage();
	async function getRandomMessage() {
		const response = await fetch('http://localhost:8080/messages/random');
		const data = await response.json();
		return data;
	}

	let message = {};

	/* notes
	const message = 'Hello world' // Try edit me

// Update header text
document.querySelector('#header').innerHTML = message

let test = "\\u00f0\\u009f\\u0098\\u0086"
let test2 = test.split("\\u")
for (let i = 1; i < test2.length; i++){
  console.log("i, val: ", i, String.fromCharCode(test2[i]))
}
let test3 = unescape(test)
console.log("test2: ", test2)
console.log("test3: ", test3)
let reg = new RegExp("([^\\u]+)")
let test4 = test.match(reg)
console.log('test4: ', test4)
// Log to console
console.log(message)
*/
	let experiment;
	async function getRandomMessage2() {
		const responseObject = await fetch('http://localhost:8080/messages/random').then((r) =>
			r.json()
		);

		console.log('responseObject: ', responseObject);

		let reactions = responseObject.Reactions;

		if (reactions) {
			let newString = reactions[0].Reaction.replace('/u', '');
			console.log('newString: ', newString);
			experiment = String.fromCodePoint(parseInt(newString));
		}

		message.content = responseObject.Content;
		message.reactions = responseObject.Reactions;
		message.sender = responseObject.Sender;
		message.timestamp = new Date(responseObject.Timestamp);
	}
	onMount(getRandomMessage2);
</script>

<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://kit.svelte.dev">kit.svelte.dev</a> to read the documentation</p>

{#await messagePromise}
	<p>Getting random message...</p>
{:then messageP}
	<p>{messageP.Sender}</p>
	<p>{messageP.Content}</p>
	<p>{messageP.Timestamp}</p>
	{#if messageP.Reactions}
		<p>Experiment: {experiment}</p>
		{#each messageP.Reactions as reaction}
			<p>Actor: {reaction.Actor}</p>
			<p>Reaction: {reaction.Reaction}</p>
		{/each}
	{/if}
{:catch error}
	<p>Oops: {error}</p>
{/await}

{#if message}
	<p>Test: {message.sender}</p>
{:else}
	<p>:(</p>
{/if}

<button on:click={() => (messagePromise = getRandomMessage())}> Explore the next planet </button>
