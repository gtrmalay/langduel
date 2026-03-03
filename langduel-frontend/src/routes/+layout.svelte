<script>
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/stores';
	import Header from '$lib/components/Header.svelte';
	import { duel } from '$lib/stores/duel.js';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<Header show={$page.url.pathname !== '/battle'} />
{@render children()}

<style>
	@import url('https://fonts.googleapis.com/css2?family=Press+Start+2P&family=Space+Grotesk:wght@400;600;700&display=swap');

	:global {
		:root {
			--bg: #0b1020;
			--panel: #111827;
			--card: #171f33;
			--text: #e9edf6;
			--muted: #9aa4b2;
			--accent: #25f4b7;
			--accent-2: #f6c144;
			--danger: #ff5c7a;
			--outline: #2b344a;
			--glow: rgba(37,244,183,0.35);
		}
		body {
			margin: 0;
			font-family: "Space Grotesk", system-ui, sans-serif;
			background:
				radial-gradient(1200px 800px at 20% 15%, rgba(37,244,183,0.12), transparent 60%),
				radial-gradient(1200px 800px at 85% 10%, rgba(246,193,68,0.08), transparent 60%),
				linear-gradient(180deg, #0a0f1f, #0b1020 60%);
			color: var(--text);
		}
		body::before {
			content: "";
			position: fixed;
			inset: 0;
			background-image:
				linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px),
				linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px);
			background-size: 36px 36px;
			opacity: 0.4;
			pointer-events: none;
			z-index: 0;
		}
		.wrap {
			max-width: 1080px;
			margin: 36px auto;
			padding: 0 24px 60px;
			position: relative;
			z-index: 1;
		}
		.app-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 16px;
			padding: 18px 24px 0;
			max-width: 1080px;
			margin: 0 auto;
		}
		.brand {
			cursor: pointer;
		}
		.logo {
			font-family: "Press Start 2P", cursive;
			font-size: 14px;
			letter-spacing: 2px;
			text-transform: uppercase;
			color: var(--text);
		}
		.nav {
			display: flex;
			align-items: center;
			gap: 10px;
			flex-wrap: wrap;
		}
		.user-pill {
			padding: 6px 10px;
			border-radius: 999px;
			border: 1px solid var(--outline);
			background: rgba(11,15,31,0.8);
			font-size: 12px;
			color: var(--muted);
		}
		header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			margin-bottom: 20px;
		}
		.title {
			font-family: "Press Start 2P", cursive;
			font-size: 22px;
			letter-spacing: 2px;
			text-transform: uppercase;
		}
		.title-badge {
			display: inline-block;
			margin-top: 10px;
			padding: 6px 12px;
			border-radius: 6px;
			font-size: 10px;
			letter-spacing: 1px;
			background: rgba(37,244,183,0.1);
			border: 1px solid rgba(37,244,183,0.5);
			color: var(--accent);
			text-transform: uppercase;
		}
		.status {
			font-size: 12px;
			color: var(--muted);
		}
		.notice {
			margin-top: 12px;
			padding: 12px 14px;
			border-radius: 8px;
			border: 1px solid var(--outline);
			background: rgba(17,24,39,0.9);
			color: var(--muted);
			font-size: 12px;
		}
		.notice.error {
			border-color: rgba(255,92,122,0.6);
			color: #ffd7df;
			background: rgba(255,92,122,0.12);
		}
		.panel {
			background: rgba(17,24,39,0.9);
			border-radius: 16px;
			padding: 28px;
			margin-top: 30px;
			border: 1px solid var(--outline);
			box-shadow: 0 18px 30px rgba(0,0,0,0.35);
			display: grid;
			gap: 20px;
		}
		.card-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
			gap: 16px;
		}
		.card-tile {
			border-radius: 12px;
			border: 1px solid rgba(37,244,183,0.25);
			background: rgba(11,15,31,0.7);
			padding: 14px;
			display: grid;
			gap: 10px;
		}
		.card-title {
			font-size: 11px;
			letter-spacing: 1px;
			text-transform: uppercase;
			color: var(--muted);
		}
		.panel h3 {
			margin: 0 0 8px 0;
			font-size: 12px;
			color: var(--muted);
			font-weight: 600;
			text-transform: uppercase;
			letter-spacing: 2px;
		}
		.section {
			padding: 18px;
			border-radius: 12px;
			border: 1px dashed rgba(37,244,183,0.25);
			background: rgba(11,15,31,0.7);
			display: grid;
			gap: 16px;
		}
		.controls {
			display: grid;
			grid-template-columns: 1fr 1fr;
			gap: 22px;
		}
		.controls.one { grid-template-columns: 1fr; }
		.controls.two { grid-template-columns: 1fr 1fr; }
		.controls.three { grid-template-columns: 1fr 1fr 1fr; }
		.controls.flow {
			grid-template-columns: minmax(200px, 1fr) auto;
			align-items: center;
		}
		input, button, select {
			padding: 12px 14px;
			border-radius: 10px;
			border: 1px solid var(--outline);
			background: rgba(11,15,31,0.9);
			color: var(--text);
		}
		input, select {
			height: 52px;
		}
		input:focus, select:focus {
			outline: 2px solid rgba(37,244,183,0.5);
			box-shadow: 0 0 0 3px rgba(37,244,183,0.15);
		}
		button {
			cursor: pointer;
			background: linear-gradient(135deg, rgba(37,244,183,0.2), rgba(37,244,183,0.05));
			border-color: rgba(37,244,183,0.5);
			color: var(--text);
			font-weight: 600;
			letter-spacing: 0.5px;
			text-transform: uppercase;
		}
		button.btn {
			height: 52px;
			min-width: 140px;
		}
		button.ghost {
			background: transparent;
			border-color: var(--outline);
			color: var(--muted);
		}
		button.secondary {
			background: rgba(11,15,31,0.9);
			border-color: var(--outline);
		}
		button:hover {
			box-shadow: 0 0 18px rgba(37,244,183,0.35);
		}
		button:active { transform: translateY(1px); }
		.small {
			font-size: 12px;
			color: var(--muted);
		}
		.status-badge {
			display: inline-flex;
			align-items: center;
			gap: 8px;
			padding: 6px 12px;
			border-radius: 999px;
			border: 1px solid var(--outline);
			background: rgba(11,15,31,0.9);
			font-size: 11px;
			color: var(--muted);
		}
		.status-dot {
			width: 8px;
			height: 8px;
			border-radius: 50%;
			background: var(--muted);
		}
		.status-dot.on { background: var(--accent); }
		.toggle {
			display: inline-flex;
			gap: 6px;
			align-items: center;
			background: rgba(11,15,31,0.9);
			border: 1px solid var(--outline);
			border-radius: 999px;
			padding: 4px;
		}
		.toggle button {
			padding: 6px 10px;
			border-radius: 999px;
			border: 1px solid transparent;
			background: transparent;
			color: var(--muted);
			font-size: 11px;
			letter-spacing: 1px;
			text-transform: uppercase;
			box-shadow: none;
		}
		.toggle button.active {
			color: var(--text);
			background: rgba(37,244,183,0.2);
			border-color: rgba(37,244,183,0.5);
		}
		.link {
			word-break: break-all;
			color: var(--muted);
		}
		.link span { color: var(--text); }
		.profile-fab {
			position: fixed;
			top: 18px;
			right: 18px;
			z-index: 40;
			border-radius: 10px;
			padding: 10px 14px;
			border: 1px solid rgba(246,193,68,0.6);
			background: rgba(246,193,68,0.12);
			color: var(--text);
			box-shadow: 0 6px 18px rgba(0,0,0,0.3);
		}
		.profile-backdrop {
			position: fixed;
			inset: 0;
			background: rgba(0,0,0,0.6);
			z-index: 45;
			border: 0;
			padding: 0;
			cursor: pointer;
		}
		.profile-panel {
			position: fixed;
			top: 70px;
			right: 18px;
			width: 330px;
			z-index: 50;
			background: rgba(17,24,39,0.95);
			border: 1px solid var(--outline);
			border-radius: 14px;
			padding: 18px;
			box-shadow: 0 12px 30px rgba(0,0,0,0.4);
		}
		.profile {
			padding: 14px;
			border-radius: 10px;
			border: 1px solid rgba(37,244,183,0.2);
			background: rgba(11,15,31,0.7);
			display: grid;
			gap: 10px;
			font-size: 12px;
			color: var(--muted);
		}
		.profile-row {
			display: flex;
			justify-content: space-between;
			gap: 12px;
		}
		.duel-list {
			display: grid;
			gap: 10px;
		}
		.duel-card {
			padding: 10px 12px;
			border-radius: 10px;
			border: 1px solid var(--outline);
			background: rgba(11,15,31,0.8);
			display: grid;
			gap: 6px;
		}
		.duel-row {
			display: flex;
			justify-content: space-between;
			gap: 10px;
			font-size: 12px;
		}
		.badge {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			padding: 2px 8px;
			border-radius: 6px;
			border: 1px solid var(--outline);
			background: rgba(11,15,31,0.8);
			font-size: 10px;
			color: var(--muted);
			text-transform: uppercase;
		}
		.badge.win { color: var(--accent); border-color: rgba(37,244,183,0.5); }
		.badge.loss { color: var(--danger); border-color: rgba(255,92,122,0.5); }
		.badge.pending { color: var(--accent-2); border-color: rgba(246,193,68,0.5); }
		.game-over {
			margin-top: 20px;
			padding: 16px;
			border-radius: 12px;
			border: 1px solid rgba(246,193,68,0.5);
			background: rgba(246,193,68,0.08);
			display: grid;
			gap: 12px;
		}
		.stats {
			margin-top: 18px;
			padding: 14px;
			border-radius: 12px;
			border: 1px solid var(--outline);
			background: rgba(11,15,31,0.7);
			display: grid;
			gap: 10px;
			font-size: 12px;
			color: var(--muted);
		}
		.stats-row {
			display: flex;
			justify-content: space-between;
			gap: 12px;
		}
		.arena {
			display: grid;
			grid-template-columns: 1fr 120px 1fr;
			gap: 34px;
			align-items: center;
			margin-top: 12px;
		}
		.card {
			background: var(--card);
			border-radius: 14px;
			padding: 18px;
			box-shadow: 0 12px 30px rgba(0,0,0,0.4);
			position: relative;
			overflow: hidden;
			border: 1px solid rgba(37,244,183,0.15);
		}
		.card::after {
			content: '';
			position: absolute;
			inset: 0;
			border: 1px solid rgba(255,255,255,0.04);
			pointer-events: none;
		}
		.player-name {
			font-family: "Press Start 2P", cursive;
			font-size: 12px;
			margin-bottom: 10px;
			text-transform: uppercase;
			letter-spacing: 2px;
		}
		.player-card {
			background: var(--card);
			border-radius: 16px;
			border: 1px solid rgba(37,244,183,0.15);
			padding: 18px;
			display: grid;
			gap: 10px;
		}
		.player-meta {
			display: flex;
			justify-content: space-between;
			font-size: 12px;
			color: var(--muted);
		}
		.hp-bar {
			height: 10px;
			border-radius: 8px;
			background: rgba(11,15,31,0.9);
			overflow: hidden;
			border: 1px solid rgba(37,244,183,0.2);
		}
		.hp-bar span {
			display: block;
			height: 100%;
			background: linear-gradient(90deg, var(--accent), #19d9a4);
		}
		.timer-pill {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			padding: 8px 14px;
			border-radius: 999px;
			border: 1px solid rgba(246,193,68,0.6);
			color: var(--accent-2);
			letter-spacing: 2px;
			font-size: 12px;
		}
		.chat {
			border-radius: 16px;
			border: 1px solid var(--outline);
			background: rgba(11,15,31,0.8);
			padding: 16px;
			display: grid;
			gap: 10px;
		}
		.chat-title {
			text-transform: uppercase;
			font-size: 12px;
			color: var(--muted);
		}
		.chat-list {
			min-height: 120px;
			max-height: 180px;
			overflow: auto;
			display: grid;
			gap: 6px;
			padding-right: 4px;
		}
		.chat-item {
			background: rgba(17,24,39,0.8);
			padding: 6px 8px;
			border-radius: 8px;
			font-size: 12px;
		}
		.chat-empty {
			color: var(--muted);
			font-size: 12px;
		}
		.chat-input {
			display: grid;
			grid-template-columns: 1fr auto;
			gap: 8px;
			align-items: center;
		}
		.chat-note {
			font-size: 11px;
			color: var(--muted);
		}
		.hp {
			height: 12px;
			border-radius: 6px;
			background: rgba(11,15,31,0.9);
			overflow: hidden;
			border: 1px solid rgba(37,244,183,0.2);
		}
		.hp > span {
			display: block;
			height: 100%;
			width: 100%;
			background: linear-gradient(90deg, var(--accent), #19d9a4);
			transition: width 0.3s ease;
		}
		.hp.low > span {
			background: linear-gradient(90deg, var(--danger), #ff8aa0);
		}
		.hp-label {
			margin-top: 8px;
			font-size: 12px;
			color: var(--muted);
		}
		.vs {
			text-align: center;
			font-family: "Press Start 2P", cursive;
			font-size: 14px;
			color: var(--accent-2);
			background: rgba(11,15,31,0.9);
			border: 1px solid rgba(246,193,68,0.6);
			padding: 12px 0;
			border-radius: 10px;
			box-shadow: inset 0 0 16px rgba(246,193,68,0.25);
		}
		.prompt {
			margin-top: 38px;
			background: rgba(17,24,39,0.95);
			border-radius: 14px;
			padding: 28px;
			text-align: center;
			font-size: 22px;
			letter-spacing: 1px;
			border: 1px solid rgba(37,244,183,0.3);
		}
		.timer {
			margin-top: 16px;
			text-align: center;
			font-size: 12px;
			letter-spacing: 2px;
			color: var(--accent-2);
			text-transform: uppercase;
		}
		.sub {
			margin-top: 16px;
			text-align: center;
			color: var(--muted);
			font-size: 12px;
		}
		.hit {
			animation: hitFlash 0.35s ease;
		}
		@keyframes hitFlash {
			0% { box-shadow: 0 0 0 rgba(255,92,122,0); }
			50% { box-shadow: 0 0 22px rgba(255,92,122,0.9); }
			100% { box-shadow: 0 0 0 rgba(255,92,122,0); }
		}
		@media (max-width: 720px) {
			.arena { grid-template-columns: 1fr; }
			.vs { display: none; }
			.controls { grid-template-columns: 1fr; gap: 18px; }
			.panel { padding: 22px; }
		}
	}
</style>
