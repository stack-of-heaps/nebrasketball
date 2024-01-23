<script lang="ts">
	import type { Message } from '$lib/types';
	export let message = {} as Message;
</script>

<div>
	<b>{message.sender}</b>
	<p class="tiny">{message.time}</p>
</div>
<p>{message.content}</p>
{#if message.reactions}
	{#each message.reactions as reaction}
		<p>{reaction.actor} {reaction.reaction}</p>
	{/each}
{/if}

<style>
	.tiny {
		font-size: 8px;
	}
</style>
