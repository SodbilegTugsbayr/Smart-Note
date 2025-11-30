<script setup>
const text = ref("")
const responses = ref([])

async function send() {
  const data = await $fetch("/pub/echo", {
    method: "POST",
    body: text.value
  })
  responses.value.push(data)
}
</script>
<template>
  <v-container>
    <v-row>
      <v-col>
        <h1>Hello!</h1>
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-text-field v-model="text" variant="outlined" color="primary">
          <template #append-inner>
            <v-btn color="primary" icon="mdi-send" variant="text" @click="send"></v-btn>
          </template>
        </v-text-field>
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <h1>Echo endpoint responses</h1>
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="12">
        <div v-for="resp of responses">
          {{ resp }}
        </div>
      </v-col>
    </v-row>
  </v-container>
</template>