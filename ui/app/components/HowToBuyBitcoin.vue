<template>
  <v-card-text>
    <v-card-title class="title-crypto">{{$t('how_to_buy.crypto_title')}}</v-card-title>
    <p>{{$t('how_to_buy.first_text')}}</p>
    <p>{{$t('how_to_buy.second_text')}}</p>
    <ul class="crypto-list">
      <li><a href="https://academy.binance.com/en/articles/binance-beginner-s-guide" target="_blank" class="crypto-link text-primary">Binance</a></li>
      <li><a href="https://www.coinbase.com/en/how-to-buy/bitcoin" target="_blank" class="crypto-link text-primary">Coinbase</a></li>
      <li><a href="https://metamask.io/buy-crypto" target="_blank" class="crypto-link text-primary">MetaMask</a></li>
    </ul>
  </v-card-text>
  <v-row style="box-sizing: border-box" class="px-6">
    <v-col cols="4">
      <a href="https://academy.binance.com/en/articles/binance-beginner-s-guide" target="_blank">
        <v-img src="/binance.png"/>
      </a>
    </v-col>
    <v-col cols="4">
      <a href="https://www.coinbase.com/en/how-to-buy/bitcoin" target="_blank">
        <v-img src="/coinbase.png"/>
      </a>
    </v-col>
    <v-col cols="4">
      <a href="https://metamask.io/buy-crypto" target="_blank">
        <v-img src="/metamask.png"/>
      </a>
    </v-col>
  </v-row>
</template>
<script lang="ts">
export default {
  name: "HowToBuyBitcoin",
}
</script>
<style scoped>
.crypto-link {
  text-decoration: none;
}
.crypto-list {
  margin-top: 10px;
  margin-left: 20px;
}
.title-crypto {
  margin-left: 0;
  padding-left: 0;
}
</style>