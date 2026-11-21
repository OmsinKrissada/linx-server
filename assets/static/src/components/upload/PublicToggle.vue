<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <Toggle
        variant="outline"
        :model-value="model"
        @update:model-value="(v) => (model = !!v)"
        :data-state="model ? 'on' : 'off'"
        v-bind="$attrs"
      >
        <PublicIcon class="text-2xl" />
        <span class="sr-only">Show in public list</span>
      </Toggle>
    </TooltipTrigger>
    <TooltipContent side="bottom">
      <div class="text-sm">Show in public list</div>
      <div class="text-xs text-muted-foreground">List the filename on the Browse page.</div>
    </TooltipContent>
  </Tooltip>
</template>

<script setup lang="ts">
import { Toggle } from "@/components/ui/toggle";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import PublicIcon from "~icons/material-symbols/public";

defineOptions({ inheritAttrs: false });

const model = defineModel<boolean>({ default: false });
</script>
