<script setup lang="ts">
const props = defineProps({
  error: Object,
})

const handleError = () => location.reload()

onMounted(() => {
  console.log(props.error)
})
</script>

<template>
  <div class="text-center mt-16">
    <h2 class="text-h3 font-weight-light">
      <template v-if="props.error?.statusCode == 400">Оролтын алдаа.</template>
      <template v-else-if="props.error?.statusCode == 404">Хуудас олдсонгүй.</template>
      <template v-else-if="props.error?.statusCode == 500">Серверийн алдаа.</template>
      <template v-else>Тодорхойгүй алдаа.</template>
    </h2>
    <!-- <h2 class="text-h1">{{ props.error?.statusCode }}</h2> -->
    <h2 class="text-h6 font-weight-regular text-info mt-4">
      <template v-if="props.error?.statusCode == 400">Та оролтын утгуудаа шалгаад дахин оролдоно уу.</template>
      <template v-else-if="props.error?.statusCode == 404">Энэ хуудас устсан эсвэл огт байхгүй гэсэн үг. Та линкээ
        шалгана уу.</template>
      <template v-else-if="props.error?.statusCode == 500">Та түр хүлээгээд дахин оролдоно уу. Үүний дараа ч болохгүй
        байвал админд мэдэгдэнэ
        үү.</template>
      <template v-else>Та түр хүлээгээд дахин оролдоно уу.<br />Дахин оролдсон ч болохгүй байгаа бол админд
        мэдэгдэнэ үү.</template>
    </h2>
    <v-btn class="mt-6" color="primary" @click="handleError"> Дахин ачааллах </v-btn>
  </div>
</template>
