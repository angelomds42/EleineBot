language-name = Português (BR)
language-flag = 🇧🇷
language-menu =
    <b>Idioma atual:</b> { $languageFlag } { $languageName }

    <b>Selecione abaixo o idioma que você quer utilizar no bot.</b>
language-changed = O idioma foi alterado com sucesso.
measurement-unit = m
start-button = Inciar uma conversa.
start-message =
    Olá <b>{ $userFirstName }</b> — Eu sou a <b>{ $botName }</b>, um bot com alguns comandos úteis e divertidos para você.

    <b>Código Fonte:</b> <a href='github.com/ruizlenato/smudgelord'>GitHub</a>
start-message-group =
    Olá, eu sou o <b>{ $botName }</b>
    Tenho várias funções interessantes. Para saber mais, clique no botão abaixo e inicie uma conversa comigo.
language-button = Idioma
help-button = ❔Ajuda
about-button =  ℹ️ Sobre
donation-button = 💵 Donation
news-channel-button = 📢 Canal
about-your-data-button = Sobre seus dados
back-button = ↩️ Voltar
denied-button-alert = Este botão não é para você.
privacy-policy-button = 🔒 Política de Privacidade
privacy-policy-group = Para acessar a política de privacidade da Eleine, <b>clique no botão abaixo.</b>
about =
    <b>— Eleine</b>
    Sou um fork do @SmudgeLordBot com recursos adicionais.

    <b>- Código Fonte Fork:</b> <a href='https://github.com/angelomds42/EleineBot'>Github</a>
    <b>- Desenvolvedor:</b> @Knotzy07x

    <b>- Código Fonte Base:</b> <a href='https://github.com/ruizlenato/SmudgeLord'>Github</a>
    <b>- Desenvolvedor:</b> @ruizlenato

    <b>💸 Contribua como o projeto original: Este fork existe graças ao trabalho do ruizlenato. Contribua para mantê-lo ativo!</b>
    • Chave Pix e Email do PayPal: <code>ruizlenato@proton.me</code>

    Se preferir contribuir de outra forma, como com cartão de crédito ou débito, toque no botão abaixo para ser redirecionado ao link de doação no Ko-Fi.
privacy-policy-private =
    <b>Política de Privacidade da Eleine.</b>

    O Eleine foi criado com o compromisso de garantir transparência e confiança aos seus usuários. 
    Agradeço pela sua confiança e estou dedicado a proteger sua privacidade.
about-your-data = 
    <b>Sobre seus dados.</b>

    <b>1. Coleta de Dados.</b>
    O bot coleta apenas informações essenciais para proporcionar uma experiência personalizada.
    <b>Os dados coletados incluem:</b>
    - <b>Informações do usuário no Telegram:</b> ID do usuário, primeiro nome, idioma e nome de usuário.
    - <b>Suas configurações na Eleine:</b> Configurações que você configurou no bot, como seu idioma e nome de usuário do LastFM, tudo fornecido pelo próprio usuário.

    <b>2. Uso de dados.</b>
    Os dados coletados pelo bot são utilizados exclusivamente para aprimorar a experiência do usuário e prestar um serviço mais eficiente.
    - <b>Suas informações de usuário do Telegram</b> são usadas para identificação e comunicação com o usuário.
    - <b>Suas configurações na Eleine</b> são usadas para integrar e personalizar os serviços do bot.

    <b>3. Compartilhamento de Dados.</b>
    Os dados coletados pelo bot não são compartilhados com terceiros, exceto quando exigido por lei. 
    Todos os seus dados são armazenados de forma segura.

    <b>Observação:</b> Suas informações de usuário do Telegram são informações públicas da plataforma. Não sabemos seus dados "reais".
help =
    Aqui estão todos os meus módulos.
    <b>Para saber mais sobre os módulos, basta clicar em seus nomes.</b>

    <b>Observação:</b>
    Alguns módulos possuem configurações adicionais em grupos.
    Para saber mais, envie <code>/config</code> em um grupo onde você é administrador.
relative-duration-seconds = { $count ->
    [one] { $count } segundo
    *[other] { $count } segundos
}
relative-duration-minutes = { $count ->
    [one] { $count } minuto
    *[other] { $count } minutos
}
relative-duration-hours = { $count ->
    [one] { $count } hora
    *[other] { $count } horas
}
relative-duration-days = { $count ->
    [one] { $count } dia
    *[other] { $count } dias
}
relative-duration-weeks = { $count ->
    [one] { $count } semana
    *[other] { $count } semanas
}
relative-duration-months = { $count ->
    [one] { $count } mês
    *[other] { $count } meses
}
afk = AFK
afk-help = 
    <b>AFK — <i>Away from Keyboard</i></b>

    <b>AFK</b> significa <b>"Longe do Teclado"</b> em português.
    É uma gíria da internet para informar que você está ausente.

    <b>— Comandos:</b>
    <b>/afk (motivo):</b> Marca você como ausente.
    <b>brb (motivo):</b> Funciona como o comando afk, mas não é necessário usar o <code>/</code>.
user-now-unavailable = <b>{ $userFirstName }</b> está agora indisponível!
user-unavailable =
    <b><a href='tg://user?id={ $userID }'>{ $userFirstName } </a></b> está <b>indisponível!</b>
    Visto pela última vez à <code>{ $duration}</code> atrás.
user-unavailable-reason = <b>Reason:</b> <code>{ $reason }</code>
now-available = <b><a href='tg://user?id={ $userID }'>{ $userFirstName }</a></b> está de volta após <code>{ $duration }</code> de ausência!
moderation = Moderaçao
moderation-help =
    <b>Moderaçao:</b>

    Esse módulo é feito para ser <b>utilizado em grupos.</b>
    Você deve ser administrador para utilizá-lo.

    <b>— Restriçoes:</b>
    <b>/banir [ID|resposta] (tempo) (revoke):</b> Bane um usuário do grupo.
    <b>/mute [ID|resposta] (tempo):</b> Silenciar um usuário do grupo.
    <b>/delete [resposta]:</b> Deletar uma mensagem.
    <b> - ID ou resposta:</b> Especifique o ID do usuário ou responda à mensagem dele.
    <b> - tempo:</b> (opcional) Defina por quanto tempo a restrição será aplicada (ex: 1h, 2d).
    <b> - revoke:</b> (opcional) Se definido, remove todas as mensagens do usuário.

    <b>— Configurações:</b>
    <b>/disable (comando):</b> Desativa o comando especificado no grupo.
    <b>/enable (comando):</b> Reativa o comando que foi previamente desativado.
    <b>/disableable:</b> Lista todos os comandos que podem ser desativados.
    <b>/disabled:</b> Exibe os comandos que estão atualmente desativados.
    <b>/config:</b> Abre um menu com opções de configurações do grupo.
config-message =
    <b>Configurações —</b> Aqui estão minhas configurações para esse grupo.
    Para saber mais, <b>clique nos botões abaixo.</b>
config-medias =
    <b>Configurações do módulo de mídias:</b>
    Para saber mais sobre o módulo <b><i>mídias</i></b>, use /help no meu chat privado.

    <b>Para saber mais sobre cada configuração, clique em seu nome..</b>
    <i>Essas configurações são específicas para este grupo, não se aplicam a outros grupos ou chats privados.</i>
caption-button = Legendas:
automatic-button = Automático:
command-enabled = O comando <code>{ $command }</code> foi ativado com sucesso.
command-already-enabled = O comando <code>{ $command }</code> já estava ativado.
enable-commands-usage =
    Especifique o comando que você deseja ativar. Para ver quais os comandos que estão atualmente desativados, utilize /disabled.

    <b>Uso:</b> <code>/enable (comando)</code>
no-disabled-commands = Não existem comandos desativados <b>neste grupo.</b>
disabled-commands = <b>Comandos desativados:</b>
disableables-commands = <b>Comandos desativáveis:</b>
command-already-disabled = O comando <code>{ $command }</code> já estava desativado.
command-disabled = O comando <code>{ $command }</code> foi desativado com sucesso.
disable-commands-usage =
    Especifique o comando que você deseja desativar. Para ver a lista de comandos desativáveis, utilize /disableable.

    <b>Uso:</b> <code>/disable (comando)</code>
command-not-deactivatable = O comando <code>{ $command }</code> <b>não pode ser desativado.</b>
medias = Mídias
medias-help =
    <b>Media Downloader</b>

    Ao compartilhar links no Telegram, alguns sites não exibem uma pré-visualização de imagem ou vídeo. 
    Esse módulo faz com que a Eleine detecte automaticamente os links dos sites suportados e envie os vídeos e imagens que estão presentes no mesmo.

    <b>Sites atualmente suportados:</b> <i>Instagram</i>, <i>TikTok</i>, <i>Twitter/X</i>, <i>Threads</i>, <i>Reddit</i>, <i>Bluesky</i>, <i>YouTube Shorts</i> e <i>Xiaohongshu (Rednote)</i>.

    <b>Observação:</b> 
    Esse módulo contém configurações adicionais para grupos. 
    Você pode desativar os downloads automáticos e as legendas em grupos.

    <b>— Comandos:</b>
    <b>/dl | /sdl (link):</b> Este comando é para quando você desabilita downloads automáticos em grupos.
    <b>/ytdl (link)</b>: Baixa vídeos do <b>YouTube</b>
    A qualidade máxima dos downloads de vídeo é 1080p. Você também pode baixar apenas o áudio do vídeo com este comando.
youtube-no-url =
    Você precisa especificar um link válido do YouTube para fazer o download.

    <b>Exemplo:</b> <code>/ytdl https://youtu.be/OjNpRbNdR7E</code>
youtube-invalid-url = Este link do YouTube é inválido ou é de um vídeo privado.
youtube-video-info =
    <b>Título:</b> { $title }
    <b>Autor:</b> { $author }
    <b>Tamanho:</b> <code>{ $audioSize }</code> (Áudio) | <code>{ $videoSize }</code> (Vídeo)
    <b>Duração:</b> { $duration }
youtube-download-video-button = Baixar vídeo
youtube-download-audio-button = Baixar áudio
video-exceeds-limit = 
    O vídeo excede o limite de { $size ->
       [1572864000] 1,5GB
       *[other] 50 MB
   }, meu máximo permitido.
downloading = Baixando...
uploading = Enviando...
youtube-error =
    <b>Ocorreu um erro ao processar o vídeo. Tente novamente mais tarde.</b>
    Se o problema persistir, entre em contato com meu desenvolvedor.
auto-help = Quando ativado, o bot baixará automaticamente fotos e vídeos de determinadas redes sociais ao detectar um link, eliminando a necessidade do comando /sdl ou /dl.
caption-help = Quando ativado, a legenda das mídias baixada pelo bot serão enviadas junto com a mídia.
no-link-provided =
    <b>Você não especificou um link ou especificou um link invalido.</b>
    Especifique um link do <b><i>Instagram</i></b>, <b><i>TikTok</i></b>, <b><i>Twitter/X</i></b>, <b><i>Threads</i></b>, <b><i>Reddit</i></b>, <b><i>Bluesky</i></b>, <b><i>YouTube Shorts</i></b> ou <b><i>Xiaohongshu (Rednote)</i></b> para que eu possa baixar a(s) mídia(s).
misc = Diversos
misc-help =
    <b>Miscellaneous</b>

    Esse módulo reúne alguns comandos úteis que não se encaixam em outras categorias específicas.

    <b>— Comandos:</b>
    <b>/clima (cidade):</b> Exibe o clima atual da cidade especifica.
    <b>/tr (origem)-(destino) (texto):</b> Traduz um texto do idioma de origem para o idioma de destino especificado.
    <i>Caso você não especifique o idioma de origem, a Eleine irá identificar automaticamente.</i>
        

    <b>Observação:</b>
    Você pode traduzir mensagens respondendo a elas com <code>/tr</code>.
    Ambos os comandos <code>/tr</code> e <code>/translate</code> funcionam da mesma forma.
translator-no-args-provided =
    Você precisa especificar o texto que deseja traduzir ou responder a uma mensagem de texto, ou uma foto com legenda.

    <b>Usage:</b> <code>/tr (?idioma) (texto para tradução)</code>
weather-no-location-provided =
    Você precisa especificar o local para o qual deseja saber as informações meteorológicas.
    
    <b>Exemplo:</b> <code>/clima Belém</code>.
weather-select-location = <b>Selecione o local que você deseja saber o clima:</b>
weather-details =
    <b>{ $localname }</b>:

    Temperatura: <code>{ $temperature } °C</code>
    Sensação térmica: <code>{ $temperatureFeelsLike } °C</code>
    Umidade do ar: <code>{ $relativeHumidity }%</code>
    Velocidade do vento: <code>{ $windSpeed } km/h</code>
stickers = Figurinhas
kanging = <code>Kangando (roubando) a figurinha...</code>
kang-no-reply-provided = Você precisa usar este comando respondendo a <i><b>uma figurinha</b></i>, <i><b>uma imagem</b></i> ou <i><b>um vídeo</b></i>.
converting-video-to-sticker = <code>Convertendo vídeo/gif para figurinha de vídeo...</code>
sticker-pack-already-exists = <code>Usando um pacote de figurinhas existente...</code>
kang-error =
    <b>Ocorreu um erro ao processar a figurinha, tente novamente.</b>
    Se o erro persistir, entre em contato com o meu desenvolvedor.
get-sticker-no-reply-provided =
    Você precisa usar este comando respondendo a uma <b>figurinha estática (imagem) ou de vídeo.
sticker-invalid-media-type = O arquivo que você respondeu não é valido, responda a uma <i><b>figurinha</b></i> (sticker), um <i><b>vídeo</b></i> ou <i><b>uma foto</b></i>.
sticker-new-pack = <code>Criando um novo pacote de figurinhas...</code>
sticker-stoled = 
    Figurinha roubada <b>com sucesso</b>
    <b>Emoji:</b> { $emoji }
sticker-view-pack = Ver pacote de figurinhas
stickers-help = 
    <b>Figurinhas — Stickers</b>

    Esse módulo contém algumas funções úteis para você gerenciar figurinhas (stickers).

    <b>— Comandos:</b>
    <b>/kang (emoji):</b> Responda a qualquer figurinha para adicioná-la ao seu pacote de figurinhas criado por mim. <b>Funciona com figurinha <i>estáticas, de vídeo e animadas.</i></b>
    <b>/getsticker:</b> Responda a uma figurinha para que eu possa enviá-la como arquivo <i>.png</i> ou <i>.gif</i>. <b>Funciona apenas com figurinhas <i>de vídeo ou estáticas.</i></b>
lastfm = Last.FM
no-lastfm-username-provided =
    Você precisa especificar seu nome de usuário last.fm para que eu possa salvar meu banco de dados.
    
    <b>Examplo:</b> <code>/setuser maozedong</code>.
invalid-lastfm-username =
    <b>Usuário do last.fm inválido</b>
    Verifique se você digitou corretamente seu nome de usuário last.FM e tente novamente.
lastfm-username-not-defined =
    <b>Você ainda não definiu seu nome de usuário do last.fm.</b>
    Use o comando /setuser para definir.
lastfm-username-saved = <b>Pronto</b>, seu nome de usuário do last.fm foi salvo!
lastfm-error =
    <b>Parece que ocorreu um erro.</b>
    O last.fm pode estar temporariamente indisponível.

    Tente novamente mais tarde. Se o problema persistir, entre em contacto com o meu desenvolvedor.
no-scrobbled-yet = 
    <b>Parece que você ainda não fez scrobble de nenhuma música no Last.fm.</b>

    Se você estiver enfrentando problemas com o Last.fm, visite last.fm/about/trackmymusic para saber como conectar sua conta ao seu aplicativo de música.
lastfm-playing = 
   <b><a href='https://last.fm/user/{ $lastFMUsername }'>{ $firstName }</a></b> { $nowplaying ->
       [true] está ouvindo
      *[false] estava ouvindo
   } pela <b>{ $playcount }ª vez</b>:
lastfm-help =
    <b>Last.FM Scobbles</b>

    <b>Scrobble</b> é um recurso do Last.fm que registra automaticamente as músicas que você está ouvindo ou ouviu para o serviço.
    <b>Para saber mais, <a href='https://www.last.fm/pt/about/trackmymusic'>clique aqui</a>.</b>

    <b>— Comandos:</b>
    <b>/setuser (nome de usuário):</b> Define seu nome de usuário do Last.fm.
    <b>/lastfm | /lp:</b> Exibe a música que você está ouvindo ou ouviu recentemente.
    <b>/album | /alb:</b> Exibe o álbum que você está ouvindo ou ouviu recentemente.
    <b>/artist   | /art:</b> Exibe o artista que você está ouvindo ou ouviu recentemente.
id-required = Você precisa responder a uma mensagem ou fornecer o ID do usuário.
id-invalid = Não foi possível encontrar esse usuário. Responda a uma mensagem ou informe um ID válido.
ban-success = O usuário <a>{ $userBannedFirstName }</a> foi banido permanentemente.
unban-success = O usuário <a>{ $userUnbannedFirstName }</a> foi desbanido foi desbanido e pode voltar ao grupo.
ban-failed = Não foi possível banir este usuário.
mute-failed = Não foi possível silenciar este usuário.
mute-success = O usuário <a>{ $userMutedFirstName }</a> foi silenciado permanentemente.
unmute-success = O usuario <a>{ $userUnmutedFirstName }</a> pode enviar mensagens novamente.
mute-success-temp = O usuário <a>{ $userMutedFirstName }</a> foi silenciado até <code>{ $untilDate }</code>.
ban-success-temp = O usuário <a>{ $userBannedFirstName }</a> foi banido até <code>{ $untilDate }</code>.
delete-msg-id-required = Você precisa responder a mensagem que deseja deletar.
delete-msg-failed = Não foi possível excluir a mensagem. Bots só podem deletar mensagens com até 48 horas de envio.
delete-msg-success = Mensagem deletada com sucesso.
bot-not-admin = Preciso ser administrador para executar este comando.
user-not-admin = Você não tem permissão para executar este comando.
device-usage-hint = Para pesquisar você precisa fornecer um nome, codinome ou modelo do dispositivo.
device-not-found = Nenhum dispositivo encontrado com o termo <code>{ $searchTerm }</code>.
device-search-error = Não foi possível buscar informações sobre esse dispositivo. Por favor, tente novamente mais tarde.
device-found = <b>Dispositivos encontrados:</b>
device-info =
    <b>Nome:</b> <code>{ $device-name }</code>
    <b>Marca:</b> <code>{ $device-brand }</code>
    <b>Modelo:</b> <code>{ $device-model }</code>
    <b>Codinome:</b> <code>{ $device-codename }</code>
device-more-results = Mais resultados disponíveis. Mostrando apenas 5 correspondências - tente uma busca mais específica.
android = Android
android-help = <b>Android:</b>

    Aqui você encontra comandos para dispositivos relacionados a dispositivos Android.
    
    <b>— Pesquisas:</b>
    <b>/device [nome|codinome|modelo]:</b> Busca informações sobre dispositivos Android.
    <b> - nome:</b> Nome comercial (ex: <i>Redmi Note 11</i>).
    <b> - codinome:</b> Codenome do dispositivo (ex: <i>spes</i>).
    <b> - modelo:</b> Modelo do dispositivo (ex: <i>2201117TG</i>).
group-only = Esse comando só pode ser usado em grupos.