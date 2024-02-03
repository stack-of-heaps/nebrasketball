<script lang="ts">
	import { Group, Chip } from '@svelteuidev/core';

	export let filterSelections = '';

	let audioSelected = false;
	let gifSelected = false;
	let photoSelected = false;
	let shareSelected = false;
	let videoSelected = false;

	function getFilters(): string {
		let filters = '';
		if (audioSelected) filters += 'audio,';
		if (gifSelected) filters += 'gifs,';
		if (photoSelected) filters += 'photos,';
		if (shareSelected) filters += 'share,';
		if (videoSelected) filters += 'videos,';

		if (filters.endsWith(',')) filters = filters.slice(0, filters.length - 1);

		return filters;
	}
</script>

<Group position="center">
	<Chip
		on:change={() => {
			audioSelected = !audioSelected;
			filterSelections = getFilters();
		}}
	>
		audio
	</Chip>
	<Chip
		on:change={() => {
			gifSelected = !gifSelected;
			filterSelections = getFilters();
		}}
	>
		gif
	</Chip>
	<Chip
		on:change={() => {
			photoSelected = !photoSelected;
			filterSelections = getFilters();
		}}
	>
		photo
	</Chip>
	<Chip
		on:change={() => {
			shareSelected = !shareSelected;
			filterSelections = getFilters();
		}}
	>
		share
	</Chip>
	<Chip
		size="xs"
		color="purple"
		on:change={() => {
			videoSelected = !videoSelected;
			filterSelections = getFilters();
		}}
	>
		video
	</Chip>
</Group>

<style>
	div {
		font-size: 8px;
	}
</style>
