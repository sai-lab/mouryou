require 'time'

$file_path = ARGV[0]
$base_log = $file_path.split("/").last
$dir = $file_path.split("/").first
$log_date = $base_log.split(".")[0]

$data = Array.new(){ Array.new() }

def guess_data(index, row)
  if index == 0
    return
  end
  i = 1
  if index + i >= $data.length
    return
  end
  while $data[index+i][row].chomp.to_i == 0
    i += 1
    if index + i >= $data.length
      return
    end
  end

  after_i = index + i
  if $data[after_i][row].chomp.to_i == 0
    puts"its 0"
    return
  end

  time_before = Time.parse($data[index][0]) - Time.parse($data[index-1][0])
  time_after = Time.parse($data[after_i][0]) - Time.parse($data[index][0])

  diff = ($data[after_i][row].chomp.to_i - $data[index-1][row].chomp.to_i) * (time_before / (time_before + time_after))
  guess_tp = $data[index-1][row].chomp.to_i + diff

  $data[index][row] = guess_tp.to_s
end

begin
  File.open($file_path) do |file|

    file.each_line do |line|
      units = []
      units = line.split(",")
      $data.push(units)
    end
  end

  # i = 0
  # $data.each do |_, v|
  #   if v.chomp.to_i == 0
  #     guess_data(i)
  #   end
  #   i += 1
  # end

  for i in 1..$data.length-1
    for j in 1..$data[i].length-1
      if $data[i][j].chomp.to_i == 0
        guess_data(i, j)
      end
    end
  end

  # arr = []
  # $data.each do |l, value|
  #   a = []
  #   a.push(l)
  #   a.push(value)
  #   arr.push(a.join(","))
  # end

  for l in 0..$data.length-1
    $data[l] = $data[l].join(",")
    if $data[l].index("\n") == -1
      $data[l] = $data[l] + "\n"
    end
  end

  File.open($dir + "/" + $log_date + "_checked.csv", "w") do |f|
    $data.each { |s| f.puts(s);}
  end

rescue SystemCallError => e
  puts %Q(class=[#{e.class}] message=[#{e.message}])
rescue IOError => e
    puts %Q(class=[#{e.class}] message=[#{e.message}])
end
