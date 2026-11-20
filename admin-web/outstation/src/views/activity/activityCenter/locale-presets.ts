import { ACTIVITY_DEPOSIT_REWARD, ACTIVITY_TYPE } from '@/api/activity';

/** 预设文案已覆盖的语种；其它语言按语种归并，匹配不到时回退英文 */
type PresetLang =
  | 'en'
  | 'zh-CN'
  | 'zh-TW'
  | 'ms'
  | 'id'
  | 'th'
  | 'vi'
  | 'pt'
  | 'es';

type LocalizedText = Record<PresetLang, string>;

/** 充值活动按充值类型区分的规则首条 */
const DEPOSIT_INTROS: Record<number, LocalizedText> = {
  [ACTIVITY_DEPOSIT_REWARD.FIRST]: {
    'en': 'Members can receive a corresponding bonus on their first deposit.',
    'zh-CN': '会员首次存款，可获得相应彩金。',
    'zh-TW': '會員首次存款，可獲得相應彩金。',
    'ms': 'Ahli boleh menerima bonus yang sepadan untuk deposit pertama.',
    'id': 'Member dapat menerima bonus yang sesuai untuk deposit pertama.',
    'th': 'สมาชิกจะได้รับโบนัสตามเงื่อนไขเมื่อฝากเงินครั้งแรก',
    'vi': 'Thành viên nạp tiền lần đầu sẽ nhận được tiền thưởng tương ứng.',
    'pt': 'Os membros recebem um bônus correspondente no primeiro depósito.',
    'es': 'Los miembros reciben un bono correspondiente en su primer depósito.',
  },
  [ACTIVITY_DEPOSIT_REWARD.SECOND]: {
    'en': 'Members can receive a corresponding bonus on their second deposit.',
    'zh-CN': '会员第二次存款，可获得相应彩金。',
    'zh-TW': '會員第二次存款，可獲得相應彩金。',
    'ms': 'Ahli boleh menerima bonus yang sepadan untuk deposit kedua.',
    'id': 'Member dapat menerima bonus yang sesuai untuk deposit kedua.',
    'th': 'สมาชิกจะได้รับโบนัสตามเงื่อนไขเมื่อฝากเงินครั้งที่สอง',
    'vi': 'Thành viên nạp tiền lần hai sẽ nhận được tiền thưởng tương ứng.',
    'pt': 'Os membros recebem um bônus correspondente no segundo depósito.',
    'es': 'Los miembros reciben un bono correspondiente en su segundo depósito.',
  },
  [ACTIVITY_DEPOSIT_REWARD.CUMULATIVE]: {
    'en': 'Members can receive a corresponding bonus when cumulative deposits meet the requirement.',
    'zh-CN': '会员累计存款达到要求，可获得相应彩金。',
    'zh-TW': '會員累計存款達到要求，可獲得相應彩金。',
    'ms': 'Ahli boleh menerima bonus yang sepadan apabila deposit terkumpul memenuhi syarat.',
    'id': 'Member dapat menerima bonus yang sesuai ketika deposit kumulatif memenuhi syarat.',
    'th': 'สมาชิกจะได้รับโบนัสตามเงื่อนไขเมื่อยอดฝากสะสมครบตามที่กำหนด',
    'vi': 'Thành viên nạp tích lũy đạt yêu cầu sẽ nhận được tiền thưởng tương ứng.',
    'pt': 'Os membros recebem um bônus correspondente quando os depósitos acumulados atingirem o requisito.',
    'es': 'Los miembros reciben un bono correspondiente cuando los depósitos acumulados cumplan el requisito.',
  },
  [ACTIVITY_DEPOSIT_REWARD.SINGLE]: {
    'en': 'Members can receive a corresponding bonus when a single deposit meets the requirement.',
    'zh-CN': '会员单笔存款达到要求，可获得相应彩金。',
    'zh-TW': '會員單筆存款達到要求，可獲得相應彩金。',
    'ms': 'Ahli boleh menerima bonus yang sepadan apabila deposit tunggal memenuhi syarat.',
    'id': 'Member dapat menerima bonus yang sesuai ketika deposit tunggal memenuhi syarat.',
    'th': 'สมาชิกจะได้รับโบนัสตามเงื่อนไขเมื่อฝากเงินครั้งเดียวครบตามที่กำหนด',
    'vi': 'Thành viên nạp một lần đạt yêu cầu sẽ nhận được tiền thưởng tương ứng.',
    'pt': 'Os membros recebem um bônus correspondente quando um único depósito atingir o requisito.',
    'es': 'Los miembros reciben un bono correspondiente cuando un depósito único cumpla el requisito.',
  },
};

/** 活动类型 → 各语言活动名称（不超过 60 个字符） */
const ACTIVITY_NAMES: Record<number, LocalizedText> = {
  [ACTIVITY_TYPE.DEPOSIT]: {
    'en': 'First Deposit Extra Bonus',
    'zh-CN': '首充额外奖励',
    'zh-TW': '首充額外獎勵',
    'ms': 'Bonus Tambahan Deposit Pertama',
    'id': 'Bonus Ekstra Deposit Pertama',
    'th': 'โบนัสพิเศษฝากครั้งแรก',
    'vi': 'Thưởng thêm nạp tiền lần đầu',
    'pt': 'Bônus Extra de Primeiro Depósito',
    'es': 'Bono Extra de Primer Depósito',
  },
  [ACTIVITY_TYPE.BET]: {
    'en': 'Rebate Rewards',
    'zh-CN': '洗码奖励',
    'zh-TW': '洗碼獎勵',
    'ms': 'Ganjaran Rebat',
    'id': 'Hadiah Rebate',
    'th': 'รางวัลคืนยอดเดิมพัน',
    'vi': 'Thưởng hoàn trả cược',
    'pt': 'Recompensas de Rebate',
    'es': 'Recompensas de Reembolso',
  },
  [ACTIVITY_TYPE.REFER]: {
    'en': 'Referral Rewards',
    'zh-CN': '推荐会员奖励',
    'zh-TW': '推薦會員獎勵',
    'ms': 'Ganjaran Rujukan Ahli',
    'id': 'Hadiah Referensi Member',
    'th': 'รางวัลแนะนำสมาชิก',
    'vi': 'Thưởng giới thiệu thành viên',
    'pt': 'Recompensas por Indicação',
    'es': 'Recompensas por Referido',
  },
  [ACTIVITY_TYPE.LUCK_WHEEL]: {
    'en': 'Lucky Wheel Rewards',
    'zh-CN': '幸运轮盘奖励',
    'zh-TW': '幸運輪盤獎勵',
    'ms': 'Ganjaran Roda Bertuah',
    'id': 'Hadiah Roda Keberuntungan',
    'th': 'รางวัลวงล้อนำโชค',
    'vi': 'Phần thưởng vòng quay may mắn',
    'pt': 'Recompensas da Roleta da Sorte',
    'es': 'Recompensas de la Ruleta de la Suerte',
  },
  [ACTIVITY_TYPE.ASSIST]: {
    'en': 'Friend Boost Rewards',
    'zh-CN': '助力奖励',
    'zh-TW': '助力獎勵',
    'ms': 'Ganjaran Bantuan Rakan',
    'id': 'Hadiah Bantuan Teman',
    'th': 'รางวัลเพื่อนช่วยปลดล็อก',
    'vi': 'Phần thưởng trợ lực',
    'pt': 'Recompensas de Ajuda de Amigos',
    'es': 'Recompensas de Ayuda de Amigos',
  },
  [ACTIVITY_TYPE.BLIND_BOX]: {
    'en': 'Mystery Box Rewards',
    'zh-CN': '盲盒奖励',
    'zh-TW': '盲盒獎勵',
    'ms': 'Ganjaran Kotak Misteri',
    'id': 'Hadiah Kotak Misteri',
    'th': 'รางวัลกล่องสุ่ม',
    'vi': 'Phần thưởng hộp bí ẩn',
    'pt': 'Recompensas da Caixa Misteriosa',
    'es': 'Recompensas de la Caja Misteriosa',
  },
  [ACTIVITY_TYPE.NEWCOMER]: {
    'en': 'New Member Bonus',
    'zh-CN': '新人彩金',
    'zh-TW': '新人彩金',
    'ms': 'Bonus Ahli Baharu',
    'id': 'Bonus Member Baru',
    'th': 'โบนัสสมาชิกใหม่',
    'vi': 'Thưởng thành viên mới',
    'pt': 'Bônus de Novo Membro',
    'es': 'Bono de Nuevo Miembro',
  },
  [ACTIVITY_TYPE.AGENT]: {
    'en': 'Agent Program Rewards',
    'zh-CN': '全民代理奖励',
    'zh-TW': '全民代理獎勵',
    'ms': 'Ganjaran Program Ejen',
    'id': 'Hadiah Program Agen',
    'th': 'รางวัลตัวแทนทุกคน',
    'vi': 'Phần thưởng đại lý toàn dân',
    'pt': 'Recompensas do Programa de Agentes',
    'es': 'Recompensas del Programa de Agentes',
  },
  [ACTIVITY_TYPE.RESCUE]: {
    'en': 'Loss Rescue Bonus',
    'zh-CN': '救援金奖励',
    'zh-TW': '救援金獎勵',
    'ms': 'Bonus Penyelamat Kerugian',
    'id': 'Bonus Penyelamat Kekalahan',
    'th': 'โบนัสช่วยเหลือยอดเสีย',
    'vi': 'Thưởng cứu trợ thua cược',
    'pt': 'Bônus de Resgate de Perdas',
    'es': 'Bono de Rescate por Pérdidas',
  },
  [ACTIVITY_TYPE.LUCKY_BET]: {
    'en': 'Lucky Bet Rewards',
    'zh-CN': '幸运注单奖励',
    'zh-TW': '幸運注單獎勵',
    'ms': 'Ganjaran Taruhan Bertuah',
    'id': 'Hadiah Taruhan Beruntung',
    'th': 'รางวัลบิลเดิมพันนำโชค',
    'vi': 'Phần thưởng vé cược may mắn',
    'pt': 'Recompensas de Aposta da Sorte',
    'es': 'Recompensas de Apuesta de la Suerte',
  },
  [ACTIVITY_TYPE.VIP]: {
    'en': 'VIP Bonus',
    'zh-CN': 'VIP奖金',
    'zh-TW': 'VIP獎金',
    'ms': 'Bonus VIP',
    'id': 'Bonus VIP',
    'th': 'โบนัส VIP',
    'vi': 'Thưởng VIP',
    'pt': 'Bônus VIP',
    'es': 'Bono VIP',
  },
  [ACTIVITY_TYPE.RED_PACKET]: {
    'en': 'Red Envelope Rain',
    'zh-CN': '红包雨奖励',
    'zh-TW': '紅包雨獎勵',
    'ms': 'Hujan Angpau',
    'id': 'Hujan Angpao',
    'th': 'ฝนอั่งเปา',
    'vi': 'Mưa lì xì',
    'pt': 'Chuva de Envelopes Vermelhos',
    'es': 'Lluvia de Sobres Rojos',
  },
  [ACTIVITY_TYPE.IMAGE_AD]: {
    'en': 'Promotional Announcement',
    'zh-CN': '图文广告活动',
    'zh-TW': '圖文廣告活動',
    'ms': 'Pengumuman Promosi',
    'id': 'Pengumuman Promosi',
    'th': 'ประกาศโปรโมชั่น',
    'vi': 'Thông báo khuyến mãi',
    'pt': 'Anúncio Promocional',
    'es': 'Anuncio Promocional',
  },
  [ACTIVITY_TYPE.SIGN_IN]: {
    'en': 'Daily Check-in Rewards',
    'zh-CN': '签到奖励',
    'zh-TW': '簽到獎勵',
    'ms': 'Ganjaran Daftar Masuk Harian',
    'id': 'Hadiah Check-in Harian',
    'th': 'รางวัลเช็คอินรายวัน',
    'vi': 'Thưởng điểm danh hằng ngày',
    'pt': 'Recompensas de Check-in Diário',
    'es': 'Recompensas de Registro Diario',
  },
};

/** 活动类型 → 各语言活动规则首条（说明玩法） */
const ACTIVITY_INTROS: Record<number, LocalizedText> = {
  [ACTIVITY_TYPE.DEPOSIT]: {
    'en': 'Members can receive a corresponding extra bonus on their first deposit.',
    'zh-CN': '会员首次存款，可获得相应额外彩金。',
    'zh-TW': '會員首次存款，可獲得相應額外彩金。',
    'ms': 'Ahli boleh menerima bonus tambahan yang sepadan untuk deposit pertama.',
    'id': 'Member dapat menerima bonus ekstra yang sesuai untuk deposit pertama.',
    'th': 'สมาชิกจะได้รับโบนัสพิเศษตามเงื่อนไขเมื่อฝากเงินครั้งแรก',
    'vi': 'Thành viên nạp tiền lần đầu sẽ nhận được tiền thưởng thêm tương ứng.',
    'pt': 'Os membros recebem um bônus extra correspondente no primeiro depósito.',
    'es': 'Los miembros reciben un bono extra correspondiente en su primer depósito.',
  },
  [ACTIVITY_TYPE.BET]: {
    'en': 'Members receive rebates in proportion to their valid bets.',
    'zh-CN': '会员根据有效投注额，可获得相应比例的洗码返水。',
    'zh-TW': '會員依有效投注額，可獲得相應比例的洗碼返水。',
    'ms': 'Ahli menerima rebat mengikut kadar jumlah taruhan sah.',
    'id': 'Member menerima rebate sesuai proporsi jumlah taruhan yang valid.',
    'th': 'สมาชิกจะได้รับเงินคืนตามสัดส่วนยอดเดิมพันที่ถูกต้อง',
    'vi': 'Thành viên nhận hoàn trả theo tỷ lệ dựa trên tổng cược hợp lệ.',
    'pt': 'Os membros recebem rebate proporcional às suas apostas válidas.',
    'es': 'Los miembros reciben un reembolso proporcional a sus apuestas válidas.',
  },
  [ACTIVITY_TYPE.REFER]: {
    'en': 'Invite friends to register and make a deposit to receive referral rewards.',
    'zh-CN': '邀请好友注册并完成充值，即可获得推荐奖励。',
    'zh-TW': '邀請好友註冊並完成儲值，即可獲得推薦獎勵。',
    'ms': 'Jemput rakan untuk mendaftar dan membuat deposit bagi menerima ganjaran rujukan.',
    'id': 'Undang teman untuk mendaftar dan melakukan deposit untuk mendapatkan hadiah referensi.',
    'th': 'ชวนเพื่อนสมัครและฝากเงินสำเร็จ รับรางวัลแนะนำทันที',
    'vi': 'Mời bạn bè đăng ký và nạp tiền để nhận thưởng giới thiệu.',
    'pt': 'Convide amigos para se cadastrar e depositar e receba recompensas por indicação.',
    'es': 'Invita a tus amigos a registrarse y depositar para recibir recompensas por referido.',
  },
  [ACTIVITY_TYPE.LUCK_WHEEL]: {
    'en': 'Members who meet the requirements can spin the lucky wheel for a chance to win rewards.',
    'zh-CN': '满足条件的会员可参与幸运轮盘抽奖，赢取丰厚奖励。',
    'zh-TW': '符合條件的會員可參與幸運輪盤抽獎，贏取豐厚獎勵。',
    'ms': 'Ahli yang layak boleh memutar roda bertuah untuk peluang memenangi ganjaran.',
    'id': 'Member yang memenuhi syarat dapat memutar roda keberuntungan untuk memenangkan hadiah.',
    'th': 'สมาชิกที่มีคุณสมบัติครบสามารถหมุนวงล้อนำโชคเพื่อลุ้นรับรางวัล',
    'vi': 'Thành viên đủ điều kiện có thể quay vòng quay may mắn để nhận thưởng.',
    'pt': 'Membros elegíveis podem girar a roleta da sorte para ganhar recompensas.',
    'es': 'Los miembros que cumplan los requisitos pueden girar la ruleta de la suerte para ganar recompensas.',
  },
  [ACTIVITY_TYPE.ASSIST]: {
    'en': 'Invite friends to help you; once the required number is reached, you can claim the reward.',
    'zh-CN': '邀请好友为你助力，达到指定助力人数即可领取奖励。',
    'zh-TW': '邀請好友為你助力，達到指定助力人數即可領取獎勵。',
    'ms': 'Jemput rakan untuk membantu anda; apabila bilangan yang ditetapkan dicapai, ganjaran boleh dituntut.',
    'id': 'Undang teman untuk membantu Anda; setelah jumlah yang ditentukan tercapai, hadiah dapat diklaim.',
    'th': 'ชวนเพื่อนมาช่วยกัน เมื่อครบจำนวนที่กำหนดก็รับรางวัลได้ทันที',
    'vi': 'Mời bạn bè trợ lực, đạt đủ số lượt trợ lực quy định là có thể nhận thưởng.',
    'pt': 'Convide amigos para ajudar; ao atingir o número exigido, você poderá resgatar a recompensa.',
    'es': 'Invita a tus amigos a ayudarte; al alcanzar el número requerido podrás reclamar la recompensa.',
  },
  [ACTIVITY_TYPE.BLIND_BOX]: {
    'en': 'Members who meet the requirements can open mystery boxes to receive random rewards.',
    'zh-CN': '满足条件的会员可开启盲盒，获得随机奖励。',
    'zh-TW': '符合條件的會員可開啟盲盒，獲得隨機獎勵。',
    'ms': 'Ahli yang layak boleh membuka kotak misteri untuk menerima ganjaran rawak.',
    'id': 'Member yang memenuhi syarat dapat membuka kotak misteri untuk mendapatkan hadiah acak.',
    'th': 'สมาชิกที่มีคุณสมบัติครบสามารถเปิดกล่องสุ่มเพื่อรับรางวัลแบบสุ่ม',
    'vi': 'Thành viên đủ điều kiện có thể mở hộp bí ẩn để nhận phần thưởng ngẫu nhiên.',
    'pt': 'Membros elegíveis podem abrir caixas misteriosas e receber recompensas aleatórias.',
    'es': 'Los miembros que cumplan los requisitos pueden abrir cajas misteriosas y recibir recompensas aleatorias.',
  },
  [ACTIVITY_TYPE.NEWCOMER]: {
    'en': 'Newly registered members can claim a welcome bonus after completing the required steps.',
    'zh-CN': '新注册会员完成指定任务后，即可领取新人彩金。',
    'zh-TW': '新註冊會員完成指定任務後，即可領取新人彩金。',
    'ms': 'Ahli yang baru mendaftar boleh menuntut bonus alu-aluan selepas melengkapkan langkah yang ditetapkan.',
    'id': 'Member yang baru mendaftar dapat mengklaim bonus sambutan setelah menyelesaikan langkah yang ditentukan.',
    'th': 'สมาชิกที่สมัครใหม่สามารถรับโบนัสต้อนรับได้หลังจากทำตามขั้นตอนที่กำหนด',
    'vi': 'Thành viên mới đăng ký hoàn thành nhiệm vụ quy định là có thể nhận thưởng chào mừng.',
    'pt': 'Novos membros podem resgatar o bônus de boas-vindas após concluir as etapas exigidas.',
    'es': 'Los nuevos miembros pueden reclamar el bono de bienvenida tras completar los pasos requeridos.',
  },
  [ACTIVITY_TYPE.AGENT]: {
    'en': 'Every member can become an agent and earn commission from the valid bets of invited members.',
    'zh-CN': '人人都可成为代理，根据下级会员的有效投注获得佣金。',
    'zh-TW': '人人都可成為代理，依下級會員的有效投注獲得佣金。',
    'ms': 'Setiap ahli boleh menjadi ejen dan memperoleh komisen daripada taruhan sah ahli jemputan.',
    'id': 'Setiap member dapat menjadi agen dan memperoleh komisi dari taruhan valid member undangan.',
    'th': 'สมาชิกทุกคนสามารถเป็นตัวแทนและรับค่าคอมมิชชันจากยอดเดิมพันที่ถูกต้องของสมาชิกที่เชิญมา',
    'vi': 'Mọi thành viên đều có thể trở thành đại lý và nhận hoa hồng từ cược hợp lệ của thành viên cấp dưới.',
    'pt': 'Todo membro pode se tornar agente e ganhar comissão sobre as apostas válidas dos membros indicados.',
    'es': 'Todo miembro puede convertirse en agente y ganar comisiones por las apuestas válidas de sus referidos.',
  },
  [ACTIVITY_TYPE.RESCUE]: {
    'en': 'Members whose net losses reach the required amount during the activity period can claim a rescue bonus.',
    'zh-CN': '活动期间内净亏损达到指定金额的会员，可领取相应救援金。',
    'zh-TW': '活動期間內淨虧損達到指定金額的會員，可領取相應救援金。',
    'ms': 'Ahli yang kerugian bersihnya mencapai jumlah yang ditetapkan dalam tempoh aktiviti boleh menuntut bonus penyelamat.',
    'id': 'Member dengan kekalahan bersih mencapai jumlah yang ditentukan selama periode acara dapat mengklaim bonus penyelamat.',
    'th': 'สมาชิกที่มียอดเสียสุทธิถึงจำนวนที่กำหนดในช่วงกิจกรรม สามารถรับโบนัสช่วยเหลือได้',
    'vi': 'Thành viên có số thua ròng đạt mức quy định trong thời gian sự kiện có thể nhận tiền cứu trợ tương ứng.',
    'pt': 'Membros cujas perdas líquidas atingirem o valor exigido durante o período podem resgatar o bônus de resgate.',
    'es': 'Los miembros cuyas pérdidas netas alcancen el monto requerido durante el periodo pueden reclamar el bono de rescate.',
  },
  [ACTIVITY_TYPE.LUCKY_BET]: {
    'en': 'Bets whose numbers match the lucky criteria will receive additional rewards.',
    'zh-CN': '注单号码符合幸运规则的会员，可获得额外奖励。',
    'zh-TW': '注單號碼符合幸運規則的會員，可獲得額外獎勵。',
    'ms': 'Taruhan yang nombornya memenuhi kriteria bertuah akan menerima ganjaran tambahan.',
    'id': 'Taruhan dengan nomor yang memenuhi kriteria keberuntungan akan mendapatkan hadiah tambahan.',
    'th': 'บิลเดิมพันที่มีเลขตรงตามเงื่อนไขนำโชค จะได้รับรางวัลเพิ่มเติม',
    'vi': 'Vé cược có mã số đáp ứng quy tắc may mắn sẽ nhận thêm phần thưởng.',
    'pt': 'Apostas cujos números atendam aos critérios da sorte receberão recompensas adicionais.',
    'es': 'Las apuestas cuyos números cumplan los criterios de la suerte recibirán recompensas adicionales.',
  },
  [ACTIVITY_TYPE.VIP]: {
    'en': 'VIP members can claim exclusive bonuses according to their VIP level.',
    'zh-CN': 'VIP 会员可根据等级领取对应的专属奖金。',
    'zh-TW': 'VIP 會員可依等級領取對應的專屬獎金。',
    'ms': 'Ahli VIP boleh menuntut bonus eksklusif mengikut tahap VIP mereka.',
    'id': 'Member VIP dapat mengklaim bonus eksklusif sesuai level VIP mereka.',
    'th': 'สมาชิก VIP สามารถรับโบนัสพิเศษตามระดับ VIP ของตน',
    'vi': 'Thành viên VIP có thể nhận tiền thưởng độc quyền theo cấp độ VIP.',
    'pt': 'Membros VIP podem resgatar bônus exclusivos de acordo com seu nível VIP.',
    'es': 'Los miembros VIP pueden reclamar bonos exclusivos según su nivel VIP.',
  },
  [ACTIVITY_TYPE.RED_PACKET]: {
    'en': 'Red envelopes rain down at scheduled times during the activity; grab them for random rewards.',
    'zh-CN': '活动期间指定时段开启红包雨，会员可参与抢红包赢取随机奖励。',
    'zh-TW': '活動期間指定時段開啟紅包雨，會員可參與搶紅包贏取隨機獎勵。',
    'ms': 'Hujan angpau berlaku pada waktu yang ditetapkan sepanjang aktiviti; rebut angpau untuk ganjaran rawak.',
    'id': 'Hujan angpao berlangsung pada waktu tertentu selama acara; rebut angpao untuk hadiah acak.',
    'th': 'ในช่วงกิจกรรมจะมีฝนอั่งเปาตามเวลาที่กำหนด ร่วมแย่งอั่งเปาเพื่อลุ้นรางวัลแบบสุ่ม',
    'vi': 'Trong thời gian sự kiện, mưa lì xì diễn ra vào khung giờ quy định, hãy tham gia giành lì xì để nhận thưởng ngẫu nhiên.',
    'pt': 'Envelopes vermelhos caem em horários definidos durante a atividade; pegue-os para ganhar recompensas aleatórias.',
    'es': 'Durante la actividad caen sobres rojos en horarios definidos; atrápalos para ganar recompensas aleatorias.',
  },
  [ACTIVITY_TYPE.IMAGE_AD]: {
    'en': 'Please refer to the promotional content above for details of this activity.',
    'zh-CN': '活动详情请参阅上方图文内容。',
    'zh-TW': '活動詳情請參閱上方圖文內容。',
    'ms': 'Sila rujuk kandungan promosi di atas untuk butiran aktiviti ini.',
    'id': 'Silakan lihat konten promosi di atas untuk detail acara ini.',
    'th': 'โปรดดูรายละเอียดกิจกรรมได้จากเนื้อหาโปรโมชั่นด้านบน',
    'vi': 'Vui lòng xem nội dung hình ảnh phía trên để biết chi tiết sự kiện.',
    'pt': 'Consulte o conteúdo promocional acima para obter os detalhes desta atividade.',
    'es': 'Consulta el contenido promocional anterior para conocer los detalles de esta actividad.',
  },
  [ACTIVITY_TYPE.SIGN_IN]: {
    'en': 'Check in daily to receive rewards; consecutive check-ins earn extra rewards.',
    'zh-CN': '每日签到即可领取奖励，连续签到可获得额外奖励。',
    'zh-TW': '每日簽到即可領取獎勵，連續簽到可獲得額外獎勵。',
    'ms': 'Daftar masuk setiap hari untuk menerima ganjaran; daftar masuk berturut-turut memberikan ganjaran tambahan.',
    'id': 'Check-in setiap hari untuk mendapatkan hadiah; check-in berturut-turut memberikan hadiah tambahan.',
    'th': 'เช็คอินทุกวันเพื่อรับรางวัล เช็คอินต่อเนื่องรับรางวัลเพิ่มเติม',
    'vi': 'Điểm danh mỗi ngày để nhận thưởng, điểm danh liên tục sẽ nhận thêm phần thưởng.',
    'pt': 'Faça check-in diariamente para receber recompensas; check-ins consecutivos rendem recompensas extras.',
    'es': 'Regístrate a diario para recibir recompensas; los registros consecutivos otorgan recompensas extra.',
  },
};

/** 各活动类型通用的规则条款 */
const COMMON_RULES: Record<PresetLang, string[]> = {
  'en': [
    'Rewards (excluding principal) require 1x valid turnover before withdrawal.',
    'Only the account holder may participate through normal manual operation; otherwise rewards may be cancelled or deducted, and the account may be frozen or blacklisted.',
    'To avoid differences in interpretation, the platform reserves the right of final interpretation of this activity.',
  ],
  'zh-CN': [
    '彩金（不含本金）需完成 1 倍有效投注方可提现。',
    '仅限账号本人进行正常的人为操作，否则将取消或扣除奖励、冻结账号，甚至列入黑名单。',
    '为避免文字理解差异，平台保留本活动的最终解释权。',
  ],
  'zh-TW': [
    '彩金（不含本金）需完成 1 倍有效投注方可提現。',
    '僅限帳號本人進行正常的人為操作，否則將取消或扣除獎勵、凍結帳號，甚至列入黑名單。',
    '為避免文字理解差異，平台保留本活動的最終解釋權。',
  ],
  'ms': [
    'Ganjaran (tidak termasuk modal) memerlukan 1x taruhan sah sebelum boleh dikeluarkan.',
    'Hanya pemilik akaun boleh menyertai melalui operasi manual biasa; jika tidak, ganjaran boleh dibatalkan atau ditolak, dan akaun boleh dibekukan atau disenaraihitamkan.',
    'Untuk mengelakkan perbezaan tafsiran, platform berhak membuat tafsiran muktamad bagi aktiviti ini.',
  ],
  'id': [
    'Hadiah (tidak termasuk modal) memerlukan 1x taruhan valid sebelum dapat ditarik.',
    'Hanya pemilik akun yang boleh berpartisipasi melalui operasi manual yang wajar; jika tidak, hadiah dapat dibatalkan atau dipotong, dan akun dapat dibekukan atau dimasukkan ke daftar hitam.',
    'Untuk menghindari perbedaan penafsiran, platform berhak atas penafsiran akhir acara ini.',
  ],
  'th': [
    'โบนัส (ไม่รวมเงินต้น) ต้องทำยอดเดิมพันที่ถูกต้อง 1 เท่าจึงจะถอนได้',
    'จำกัดเฉพาะเจ้าของบัญชีที่ดำเนินการด้วยตนเองตามปกติเท่านั้น มิฉะนั้นรางวัลอาจถูกยกเลิกหรือหักคืน และบัญชีอาจถูกระงับหรือขึ้นบัญชีดำ',
    'เพื่อหลีกเลี่ยงความเข้าใจที่คลาดเคลื่อน แพลตฟอร์มขอสงวนสิทธิ์ในการตีความขั้นสุดท้ายของกิจกรรมนี้',
  ],
  'vi': [
    'Tiền thưởng (không gồm vốn gốc) cần đạt 1 lần cược hợp lệ mới được rút.',
    'Chỉ chủ tài khoản mới được tham gia bằng thao tác thủ công thông thường, nếu không phần thưởng có thể bị hủy hoặc khấu trừ, tài khoản có thể bị đóng băng hoặc đưa vào danh sách đen.',
    'Để tránh khác biệt trong cách hiểu, nền tảng giữ quyền giải thích cuối cùng đối với sự kiện này.',
  ],
  'pt': [
    'O bônus (excluindo o principal) exige 1x de apostas válidas antes do saque.',
    'Somente o titular da conta pode participar por meio de operação manual normal; caso contrário, a recompensa poderá ser cancelada ou deduzida e a conta poderá ser congelada ou incluída na lista negra.',
    'Para evitar diferenças de interpretação, a plataforma reserva-se o direito de interpretação final desta atividade.',
  ],
  'es': [
    'El bono (sin incluir el capital) requiere 1x de apuestas válidas antes de retirarse.',
    'Solo el titular de la cuenta puede participar mediante operaciones manuales normales; de lo contrario, la recompensa podrá ser cancelada o deducida y la cuenta podrá ser congelada o incluida en la lista negra.',
    'Para evitar diferencias de interpretación, la plataforma se reserva el derecho de interpretación final de esta actividad.',
  ],
};

/** 把接口语言码（如 zh-HK、ms-MY、en-US）归并到已有预设的语种 */
export function resolvePresetLang(lang: string): PresetLang {
  const code = lang.toLowerCase();
  if (code.startsWith('zh')) {
    return /hant|tw|hk|mo/.test(code) ? 'zh-TW' : 'zh-CN';
  }
  const base = code.split(/[-_]/)[0];
  if (base === 'in') return 'id';
  const supported: PresetLang[] = ['en', 'ms', 'id', 'th', 'vi', 'pt', 'es'];
  return supported.find((item) => item === base) ?? 'en';
}

/**
 * 按活动类型和语言生成预填的活动名称与规则（规则为 Quill 可用的 HTML）；
 * 充值活动可再按充值类型切换规则首条；没有预设文案的类型名称留空，规则只含通用条款
 */
export function getActivityLocalePreset(
  type: number,
  lang: string,
  rechargeType?: number
) {
  const presetLang = resolvePresetLang(lang);
  const intro =
    type === ACTIVITY_TYPE.DEPOSIT && rechargeType != null
      ? DEPOSIT_INTROS[rechargeType]?.[presetLang] ??
        ACTIVITY_INTROS[type]?.[presetLang]
      : ACTIVITY_INTROS[type]?.[presetLang];
  const rules = [...(intro ? [intro] : []), ...COMMON_RULES[presetLang]];
  return {
    name: ACTIVITY_NAMES[type]?.[presetLang] ?? '',
    content: rules.map((line) => `<p>${line}</p>`).join(''),
  };
}
