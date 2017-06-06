$file_path = ARGV[0]
$base_log = $file_path.split("/").last
$dir = $file_path.split("/").first
$log_date = $base_log.split(".")[0]

$ors = []
$crs = []
$tps = []
$dls = []
$mps = []
$times = []
$critical = []
$we = []

def write_log(arr, str)
  File.open($dir + "/" + $log_date + "_" + str + ".csv", "w") do |f|
    arr.each { |s| f.puts(s);}
  end
end

def sort_log(arr, str)
  tmp = Array.new(arr.length)

  arr.each do |line|
    items = line.split(",")
    len = items.length
    sorted = Array.new(len)

    for i in 0..len-1 do
      if items[i] == nil
        next
      end

      if i == 0
        sorted[i] = items[i]
        next
      elsif i == 1
        next
      end
      br_o_i = items[i].index("[")
      br_c_i = items[i].index("]")
      id = items[i][br_o_i + 1 .. br_c_i - 1]
      sorted[id.to_i + 1] = items[i][br_c_i + 1 .. -1].chomp
    end

    sorted.each do |x|
      if x == nil
        sorted.delete(x)
      end
    end

    s = sorted.join(",")
    tmp.push(s)
  end
  write_log(tmp, str)
end

def sort_dls
  # todo
end


begin
  File.open($file_path) do |file|
    file.each_line do |line|
      case line.split(",")[1]
      when "ors" then $ors.push(line)
      when "crs" then $crs.push(line)
      when "tps" then $tps.push(line)
      when "dls" then $dls.push(line)
      when "mps" then $mps.push(line)
      when "times" then $times.push(line)
      when "critical" then $critical.push(line)
      when "we" then $we.push(line)
      end
    end
  end

  sort_log($ors, "ors")
  sort_log($crs, "crs")
  sort_log($tps, "tps")
  # sort_log($dls, "dls")
  sort_log($mps, "mps")
  sort_log($times, "times")
  write_log($critical, "critical")
  write_log($we, "we")

rescue SystemCallError => e
  puts %Q(class=[#{e.class}] message=[#{e.message}])
rescue IOError => e
    puts %Q(class=[#{e.class}] message=[#{e.message}])
end
