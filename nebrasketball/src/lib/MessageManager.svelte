<script lang="ts">
	import mapToMessage from '$lib/mapper';
	import { onMount } from 'svelte';
	import FilterComponent from './FilterComponent.svelte';
	import MessageComponent from './MessageComponent.svelte';
	import SenderComponent from './SenderComponent.svelte';
	import type { Message, ServerMessage } from '$lib/types';
	import { Center, Stack, Button, Group } from '../../node_modules/@svelteuidev/core';

	let senderSelections = '';
	$: currentSenders = senderSelections;
	let filterSelections = '';
	$: currentFilters = filterSelections;

	let contextLoaded = false;
	let loading = true;
	let messages = [] as Message[];
	let seedMessage = {} as Message;

	async function getSeedMessage(): Promise<void> {
		loading = true;
		await fetch(
			`http://localhost:8080/messages/random?participants=${currentSenders}&filters=${currentFilters}`
		)
			.then((r) => r.json())
			.then((json) => (seedMessage = mapToMessage(json)))
			.catch((err) => console.log('Error getting random message: ', err));

		messages = [seedMessage];
		messages = messages;
		loading = false;
	}

	async function getContext(timestamp: number) {
		contextLoaded = false;
		console.log('Timestamp for context: ', timestamp);

		const response: ServerMessage[] = await fetch(`http://localhost:8080/messages/${timestamp}`)
			.then((r) => r.json())
			.catch((err) => console.log('Error getting context: ', err));

		console.log('context response: ', response);
		for (let i in response) {
			i = i;
			if (response[i].timestamp === timestamp) continue;
			else messages.push(mapToMessage(response[i]));
		}

		messages = messages.sort(
			(firstMessage, secondMessage) => firstMessage.timestamp - secondMessage.timestamp
		);
		contextLoaded = true;
	}

	onMount(() => {
		getSeedMessage();
	});
</script>

<Center>
	<Stack>
		{#if loading}
			<p>Getting random message...</p>
		{/if}

		{#if !loading}
			{#each messages as message}
				<MessageComponent {message} />
			{/each}
		{/if}

		<SenderComponent bind:senderSelections />
		<FilterComponent bind:filterSelections />

		<Group position="center">
			<Button on:click={getSeedMessage}>Get Random Message</Button>
			<Button disabled={!seedMessage.timestamp} on:click={() => getContext(seedMessage.timestamp)}>
				Get Context
			</Button>
		</Group>
	</Stack>
</Center>
