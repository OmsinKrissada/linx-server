<template>
  <Alert variant="warning" role="status">
    <WarningIcon />
    <AlertTitle>Anyone can find and open this file.</AlertTitle>
    <AlertDescription>
      It will be listed on the Browse page. Set a password to keep its contents private.
    </AlertDescription>
  </Alert>
</template>

<script setup lang="ts">
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import WarningIcon from "~icons/material-symbols/warning-rounded";
</script>
