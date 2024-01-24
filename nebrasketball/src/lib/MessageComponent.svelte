<script lang="ts">
	import type { Message } from '$lib/types';
	import { Image, Paper } from '../../node_modules/@svelteuidev/core';
	export let message = {} as Message;
</script>

<Paper>
	<div>
		<b>{message.sender}</b>
		<p class="tiny">{message.time}</p>
	</div>
	{#if message.photos}
		<p>Photo URI: {message.photos[0].uri}</p>
		<Image
			fit="contain"
			src={`C:\\Users\\Kyle Karthauser\\Documents\\nebrasketball2\\nebrasketball\\static${message.photos[0].uri}`}
		/>
	{/if}
	{#if message.gifs}
		<p>GIF URI: {message.gifs[0].uri}</p>
		<Image
			fit="contain"
			src={`C:\\Users\\Kyle Karthauser\\Documents\\nebrasketball2\\nebrasketball\\static${message.gifs[0].uri}`}
		/>
	{/if}

	<p>{message.content}</p>
	{#if message.reactions}
		{#each message.reactions as reaction}
			<p>{reaction.actor} {reaction.reaction}</p>
		{/each}
	{/if}
</Paper>

<style>
	.tiny {
		font-size: 8px;
	}
</style>
