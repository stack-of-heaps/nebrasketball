<script lang>
	import { onMount } from 'svelte';

	let loading = true;
	let message = {};
	async function getRandomMessage() {
		loading = true;
		const responseObject = await fetch('http://localhost:8080/messages/random').then((r) =>
			r.json()
		);

		let timestampAsDate = new Date(responseObject.Timestamp);
		let messageDate = timestampAsDate.toDateString();
		console.log(responseObject);
		console.log(timestampAsDate.getHours());
		let timeStr = timestampAsDate.toTimeString();
		message.timestamp = responseObject.Timestamp;
		message.time = `${messageDate} @ ${timeStr.slice(0, 5)}`;
		message.content = responseObject.Content;
		message.reactions = responseObject.Reactions;
		message.sender = responseObject.Sender;
		message.reactions = responseObject.Reactions;
		loading = false;
	}

	let context = [];
	async function getContext(timestamp) {
		const response = await fetch(`http://localhost:8080/context/${timestamp}`).then((r) =>
			r.json()
		);

		console.log('context response: ', response);
		for (let entry in response) {
			context.push(entry);
		}
	}

	onMount(getRandomMessage);
</script>

{#if loading}
	<p>Getting random message...</p>
{:else}
	<p>{message.sender}</p>
	<p>{message.content}</p>
	<p>{message.time}</p>
	{#if message.reactions}
		{#each message.reactions as reaction}
			<p>{reaction.Actor} {reaction.Reaction}</p>
		{/each}
	{/if}
{/if}

{#if context.length > 0}
	<p>Context</p>
	{#each context as c}
		<p>{c}</p>
	{/each}
{/if}

<button on:click={() => getRandomMessage()}> Get Random Message </button>
<button disabled={message.timestamp !== undefined} on:click={() => getContext(message.timestamp)}>
	Get Context
</button>
